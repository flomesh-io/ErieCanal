/*
 * Copyright 2022 The flomesh.io Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
((
    ingress = pipy.solve('ingress.js'),
    upstreamMapIssuingCA = {},
    upstreamIssuingCAs = [],
    addUpstreamIssuingCA = (ca) => (
      (md5 => (
        md5 = '' + algo.hash(ca),
        !upstreamMapIssuingCA[md5] && (
          upstreamIssuingCAs.push(new crypto.Certificate(ca)),
          upstreamMapIssuingCA[md5] = true
        )
      ))()
    ),
    balancers = {
      'round-robin': algo.RoundRobinLoadBalancer,
      'least-work': algo.LeastWorkLoadBalancer,
      'hashing': algo.HashingLoadBalancer,
    },
    services = (
      Object.fromEntries(
        Object.entries(ingress.services).map(
          ([k, v]) =>(
            ((targets, balancer, balancerInst) => (
              targets = v?.upstream?.endpoints?.map?.(ep => `${ep.ip}:${ep.port}`),
              v?.upstream?.sslCert?.ca && (
                addUpstreamIssuingCA(v.upstream.sslCert.ca)
              ),
              balancer = balancers[v?.balancer || 'round-robin'] || balancers['round-robin'],
              balancerInst = new balancer(targets || []),

              [k, {
                balancer: balancerInst,
                cache: v?.sticky && new algo.Cache(
                  () => balancerInst.next()
                ),
                upstreamSSLName: v?.upstream?.sslName || null,
                upstreamSSLVerify: v?.upstream?.sslVerify || false,
                cert: v?.upstream?.sslCert?.cert,
                key: v?.upstream?.sslCert?.key
              }]
            ))()
          )
        )
      )
    ),

  ) => pipy({
    _target: undefined,
    _service: null,
    _serviceSNI: null,
    _serviceVerify: null,
    _serviceCertChain: null,
    _servicePrivateKey: null,
    _connectTLS: null,

    _serviceCache: null,
    _targetCache: null,

    _sourceIP: null,

    _g: {
      connectionID: 0,
    },

    _connectionPool: new algo.ResourcePool(
      () => ++_g.connectionID
    ),
    
    _select: (service, key) => (
      service?.cache && key ? (
        service?.cache?.get(key)
      ) : (
        service?.balancer?.next()
      )
    ),
  })

  .import({
    __route: 'main',
  })

  .pipeline()
    .handleStreamStart(
      () => (
        _serviceCache = new algo.Cache(
          // k is a balancer, v is a target
          (k) => _select(k, _sourceIP),
          (k, v) => k.balancer.deselect(v),
        ),
        _targetCache = new algo.Cache(
          // k is a target, v is a connection ID
          (k) => _connectionPool.allocate(k),
          (k, v) => _connectionPool.free(v),
        )
      )
    )
    .handleStreamEnd(
      () => (
        _targetCache.clear(),
        _serviceCache.clear()
      )
    )
    .link('outbound-http')

  .pipeline('outbound-http')
    .handleMessageStart(
      (msg) => (
        _sourceIP = __inbound.remoteAddress,
        _service = services[__route],
        _service && (
          _serviceSNI = _service?.upstreamSSLName,
          _serviceVerify = _service?.upstreamSSLVerify,
          _serviceCertChain = _service?.cert,
          _servicePrivateKey = _service?.key,
          _target = _serviceCache.get(_service)
        ),
        _connectTLS = Boolean(_serviceCertChain) && Boolean(_servicePrivateKey),

        console.log("[balancer] _sourceIP", _sourceIP),
        console.log("[balancer] _connectTLS", _connectTLS),
        console.log("[balancer] _target.id", (_target || {id : ''}).id)
      )
    )
    .branch(
      () => Boolean(_target) && !Boolean(_connectTLS), (
        $=>$.muxHTTP(() => _targetCache.get(_target)).to(
          $=>$.connect(() => _target.id)
        )
      ), () => Boolean(_target) && Boolean(_connectTLS), (
        $=>$.muxHTTP(() => _targetCache.get(_target)).to(
          $=>$.connectTLS({
            certificate: () => ({
              cert: new crypto.Certificate(_serviceCertChain),
              key: new crypto.PrivateKey(_servicePrivateKey),
            }),
            trusted: upstreamIssuingCAs,
            sni: () => _serviceSNI || undefined,
            verify: (ok, cert) => (
              !_serviceVerify && (ok = true),
              ok
            )
          }).to(
            $=>$.connect(() => _target.id)
          )
        )
      ), (
        $=>$.chain()
      )
    )
)()
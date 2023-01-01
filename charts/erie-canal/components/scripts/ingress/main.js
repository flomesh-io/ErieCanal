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

(({
    config,
    certificates,
    issuingCAs
  } = pipy.solve('config.js'),

  ) =>

  pipy({
    _passthroughTarget: undefined,
  })
  .export('main', {
    __route: undefined,
    __isTLS: false,
  })

  .listen(
    config?.http?.enabled
      ? (config?.http?.listen ? config.http.listen : 8000)
      : 0
  ).link('inbound-http')

  .listen(
    config?.tls?.enabled
      ? (config?.tls?.listen ? config.tls.listen : 8443)
      : 0
  ).link(
    'passthrough', () => config?.sslPassthrough?.enabled === true,
    'inbound-tls'
  )

  .pipeline('inbound-tls')
    .onStart(
      () => (
        (() => (
          void(__isTLS = true)
        ))()
      )
    )
    .acceptTLS({
      certificate: (sni, cert) => (
        console.log('SNI', sni),
        (sni && (
          Object.entries(certificates).find(
            ([k, v]) => (
              v?.isWildcardHost ? false : (k === sni)
            )?.[1]
          )
          ||
          Object.entries(certificates).find(
            ([k, v]) => (
              v?.isWildcardHost ? (Boolean(v?.regex) ? v.regex.test(sni) : k === sni) : false
            )?.[1]
          )
        )) || (
          config?.tls?.certificate && config?.tls?.certificate?.cert && config?.tls?.certificate?.key
            ? {
              cert: new crypto.Certificate(config.tls.certificate.cert),
              key: new crypto.PrivateKey(config.tls.certificate.key),
            }
            : undefined
        )
      ),
      trusted: Boolean(config?.tls?.mTLS) ? issuingCAs : undefined,
      verify: (ok, cert) => (
        ok
      )
    }).to('inbound-http')

  .pipeline('passthrough')
    .handleTLSClientHello(
      hello => (
        _passthroughTarget = hello.serverNames ? hello.serverNames[0] || '' : ''
      )
    )
    .branch(
      () => Boolean(_passthroughTarget), (
        $=>$.connect(() => `${_passthroughTarget}:${config.sslPassthrough.upstreamPort}`)
      ),
      (
        $=>$.replaceStreamStart(new StreamEnd)
      )
    )

  .pipeline('inbound-http')
    .demuxHTTP().to(
      $=>$.chain(config.plugins)
    )
)()

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

(config =>

  pipy({
    _passthroughTarget: undefined,
  })

  .export('main', {
      __turnDown: false,
      __isTLS: false,
    })

  .listen(config.listen)
    .link('tls-offloaded')

  .listen(config.listenTLS)
    .link(
      'passthrough', () => config.sslPassthrough.enabled === true,
      'offload'
    )

  .pipeline('offload')
    .handleStreamStart(
      () => __isTLS = true
    )
    .acceptTLS('tls-offloaded', {
      certificate: config.listenTLS && config.certificates && config.certificates.cert && config.certificates.key ? {
        cert: new crypto.CertificateChain(config.certificates.cert),
        key: new crypto.PrivateKey(config.certificates.key),
      } : undefined,
    })

  .pipeline('passthrough')
    .handleTLSClientHello(
      hello => (
        _passthroughTarget = hello.serverNames ? hello.serverNames[0] || '' : ''
      )
    )
    .branch(
      () => (_passthroughTarget !== ''), (
        $=>$.connect(() => `${_passthroughTarget}:${config.sslPassthrough.upstreamPort}`)
      ),
      () => (_passthroughTarget === ''), (
        $=>$.replaceStreamStart(new StreamEnd)
      )
    )

  .pipeline('tls-offloaded')
    .use(config.plugins, 'session')
    .demuxHTTP('request')

  .pipeline('request')
    .use(
      config.plugins,
      'request',
      'response',
      () => __turnDown
    )

)(JSON.decode(pipy.load('config/main.json')))

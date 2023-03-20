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
    {
      config,
      tlsDomains,
      tlsWildcardDomains
    } = pipy.solve('config.js'),

  ) => pipy({
    _reject: false
  })

  .import({
    __route: 'main',
    __isTLS: 'main'
  })

  .pipeline()
    .handleMessageStart(
      msg => (
        ((host, hostname)  => (
          host = msg.head.headers['host'],
          hostname = host ? host.split(":")[0] : '',

          console.log("[reject-http] hostname", hostname),
          console.log("[reject-http] __isTLS", __isTLS),

          !__isTLS && (
            _reject = (
              Boolean(tlsDomains.find(domain => domain === hostname)) ||
              Boolean(tlsWildcardDomains.find(domain => domain.test(hostname)))
            )
          ),
          console.log("[reject-http] _reject", _reject)
        ))()
      )
    )
    .branch(
      () => (_reject), (
        $ => $
          .replaceMessage(
            new Message({
              "status": 403,
              "headers": {
                "Server": "pipy/0.90.0"
              }
            }, 'Forbidden')
          )
      ), (
        $=>$.chain()
      )
    )
)()

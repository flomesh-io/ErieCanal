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
    config = pipy.solve('ingress.js'),
    router = new algo.URLRouter(
      Object.fromEntries(
        Object.entries(config.routes).map(
          ([k, { service, rewrite }]) => [
            k, { service, rewrite: rewrite && [new RegExp(rewrite[0]), rewrite[1]] }
          ]
        )
      )
    ),

  ) => pipy()

    .import({
      __route: 'main',
    })

    .pipeline()
      .handleMessageStart(
        msg => (
          ((
            r = router.find(
              msg.head.headers.host,
              msg.head.path,
            )
          ) => (
            __route = r?.service,
            r?.rewrite && (
              msg.head.path = msg.head.path.replace(r.rewrite[0], r.rewrite[1])
            ),
            console.log('[router] Request Host: ', msg.head.headers['host']),
            console.log('[router] Request Path: ', msg.head.path)
          ))()
        )
      )
      .chain()

)()

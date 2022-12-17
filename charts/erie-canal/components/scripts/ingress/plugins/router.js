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
    _router:
      Object.fromEntries(
        Object.entries(config.routes).map(
          ([k, v]) => [
            k,
            {
              ...v,
              rewrite: v.rewrite ? [
                new RegExp(v.rewrite[0]),
                v.rewrite[1],
              ] : undefined,
              matches: v.matches?.length > 0 ? (
                v.matches.map(
                  ({ type, reg }) => (
                    {
                      "type": type,
                      "reg": new RegExp(reg)
                    }
                  )
                )
              ) : null
            }
          ]
        )
      ),

    _route: null,
  })

  .export('router', {
      __service: '',
      __ingressID: '',
      __projectID: '',
      __apiID: '',
      __requestServiceCount: new stats.Counter(
        'req_service_cnt',
        ['serviceid']
      ),
      __requestIngressCount: new stats.Counter(
        'req_ingress_cnt',
        ['ingressid']
      ),
      __requestProjectCount: new stats.Counter(
        'req_project_cnt',
        ['projectid']
      ),
    })

  .pipeline('request')
    .handleMessageStart(
      msg => (
        _route = new algo.URLRouter(_router).find(
          msg.head.headers.host,
          msg.head.path,
        ),
        _route || (
          Object.entries(_router).map(
            ([k, v]) => (
              k.split('/')[0].length > 0 ? (msg.head.headers.host ? k.split('/')[0] === msg.head.headers.host : msg.head.headers.host) :
                v.matches && (_route = v.matches.find(
                  match => (
                    match.type == 'header' && msg.head.headers[match.name] && match.reg.test(msg.head.headers[match.name]) // TODO: check other match types
                    || match.type == 'path' && msg.head.path && match.reg.test(msg.head.path)
                    || match.type == 'method' && msg.head.method && match.reg.test(msg.head.method)
                  )
                ) ? v : _route)
            )
          )
        ),
        _route && (
          __service = _route.service,
          __ingressID = _route.ingressId,
          __projectID = _route.projectId,
          __apiID = _route.serviceId,
          __requestServiceCount.withLabels(__service).increase(),
          __requestIngressCount.withLabels(__ingressID).increase(),
          __requestProjectCount.withLabels(__projectID).increase(),
          _route.rewrite && (
            msg.head.path = msg.head.path.replace(
              _route.rewrite[0],
              _route.rewrite[1],
            )
          )
        ),
        console.log('Request Host: ', msg.head.headers['host']),
        console.log('Request Path: ', msg.head.path)
      )
    )

)(JSON.decode(pipy.load('config/router.json')))

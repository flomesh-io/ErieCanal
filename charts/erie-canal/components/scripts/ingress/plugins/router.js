/*
 * MIT License
 *
 * Copyright (c) since 2021,  flomesh.io Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
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

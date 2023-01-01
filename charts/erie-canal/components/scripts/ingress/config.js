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
(
  (
    config = JSON.decode(pipy.load('config/main.json')),
    ingress = pipy.solve('ingress.js'),
    global
  ) => (
    global = {
      mapIssuingCA: {},
      issuingCAs: [],
      mapTLSDomain: {},
      tlsDomains: [],
      mapTLSWildcardDomain: {},
      tlsWildcardDomains: [],
      certificates: {}
    },

    global.addIssuingCA = ca => (
      (md5 => (
        md5 = '' + algo.hash(ca),
        !global.mapIssuingCA[md5] && (
          global.issuingCAs.push(new crypto.Certificate(ca)),
            global.mapIssuingCA[md5] = true
        )
      ))()
    ),

    global.addTLSDomain = domain => (
      (md5 => (
        md5 = '' + algo.hash(domain),
        !global.mapTLSDomain[md5] && (
          global.tlsDomains.push(domain),
            global.mapTLSDomain[md5] = true
        )
      ))()
    ),

    global.addTLSWildcardDomain = domain => (
      (md5 => (
        md5 = '' + algo.hash(domain),
        !global.mapTLSWildcardDomain[md5] && (
          global.tlsWildcardDomains.push(global.globStringToRegex(domain)),
            global.mapTLSWildcardDomain[md5] = true
        )
      ))()
    ),

    global.prepareQuote = (str, delimiter) => (
      (
        () => (
          (str + '')
          .replace(
            new RegExp('[.\\\\+*?\\[\\^\\]$(){}=!<>|:\\' + (delimiter || '') + '-]', 'g'),
'\\$&'
          )
        ))()
    ),

    global.globStringToRegex = (str) => (
      (
        () => (
          new RegExp(
            global.prepareQuote(str)
            .replace(
              new RegExp('\\\*', 'g'), '.*')
                .replace(new RegExp('\\\?', 'g'),
  '.'
            ),
      'g'
          )
        ))()
    ),

    global.issuingCAs && (
      Object.values(ingress.certificates).forEach(
        (v) => (
          v?.certificate?.ca && (
            global.addIssuingCA(v.certificate.ca)
          )
        )
      )
    ),

    global.issuingCAs && (
      ingress?.trustedCAs?.forEach(
        (v) => (
          global.addIssuingCA(v)
        )
      )
    ),

    config?.tls?.certificate?.ca && (
      global.addIssuingCA(config.tls.certificate.ca)
    ),

    global.certificates = (
      Object.fromEntries(
        Object.entries(ingress.certificates).map(
          ([k, v]) =>(
            (() => (
              v?.isTLS && Boolean(k) && (
                v?.isWildcardHost ? global.addTLSWildcardDomain(k) : global.addTLSDomain(k)
              ),
              
              [k, {
                isTLS: v?.isTLS || false,
                verifyClient: v?.verifyClient || false,
                verifyDepth: v?.verifyDepth || 1,
                cert: v?.certificate?.cert
                  ? new crypto.Certificate(v.certificate.cert)
                  : (
                    config?.tls?.certificate?.cert
                      ? new crypto.Certificate(config.tls.certificate.cert)
                      : undefined
                  ),
                key: v?.certificate?.key
                  ? new crypto.PrivateKey(v.certificate.key)
                  : (
                    config?.tls?.certificate?.key
                      ? new crypto.PrivateKey(config.tls.certificate.key)
                      : undefined
                  ),
                regex: v?.isWildcardHost ? global.globStringToRegex(k) : undefined
              }]
            ))()
          )
        )
      )
    ),

    global.config = config,

    global
  )
)()
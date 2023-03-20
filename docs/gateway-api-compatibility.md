# Gateway API Compatibility

This document describes which Gateway API resources ErieCanal supports and the extent of that support.

## Summary

| Resource | Support Status |
|-|-|
| [GatewayClass](#gatewayclass) | Partially supported |
| [Gateway](#gateway) | Partially supported |
| [HTTPRoute](#httproute) | Partially supported |
| [TLSRoute](#tlsroute) | Not supported |
| [TCPRoute](#tcproute) | Not supported |
| [UDPRoute](#udproute) | Not supported |
| [ReferencePolicy](#referencepolicy) |  Not supported |
| [ReferenceGrant](#referencegrant) |  Not supported |
| [Custom policies](#custom-policies) | Not supported |


## Terminology

We use the following words to describe support status:
- *Supported*. The resource or field is fully supported and conformant to the Gateway API specification.
- *Partially supported*. The resource or field is supported partially or with limitations. It will become fully supported in future releases.
- *Not supported*. The resource or field is not yet supported. It will become partially or fully supported in future releases.

## Resources

Below we list the resources and the support status of their corresponding fields. 

For a description of each field, visit the [Gateway API documentation](https://gateway-api.sigs.k8s.io/references/spec/). 

### GatewayClass 

> Status: Partially supported. 

ErieCanal supports only GatewayClass resource whose ControllerName is `flomesh.io/gateway-controller`. If multiple valid GatewayClasses are created, the oldest is active and take effect.


Fields:
- `spec`
	- `controllerName` - supported.
	- `parametersRef` - not supported.
	- `description` - supported.
- `status`
	- `conditions` - partially supported. Added ConditionType `Active`.

### Gateway

> Status: Partially supported.

ErieCanal supports only a single Gateway resource per namespace. 
The Gateway resource must reference ErieCanal's corresponding effective GatewayClass. 
In case of multiple Gateway resources created in the same namespace, ErieCanal will choose the oldest ONE by creation timestamp. If the timestamps are equal, ErieCanal will choose the resource that appears first in alphabetical order by “{name}”. We might support multiple Gateway resources. 

Fields:
- `spec`
	* `gatewayClassName` - supported.
	* `listeners`
		* `name` - supported.
		* `hostname` - partially supported. Wildcard hostnames like `*.example.com` are not yet supported.
		* `port` - partially supported. Allowed values: `80` for HTTP listeners and `443` for HTTPS listeners.
		* `protocol` - partially supported. Allowed values: `HTTP`, `HTTPS`.
		* `tls`
		  * `mode` - partially supported. Allowed value: `Terminate`.
		  * `certificateRefs` - partially supported. The TLS certificate and key must be stored in a Secret resource of type `kubernetes.io/tls` in the same namespace as the Gateway resource. Only a single reference is supported. You must deploy the Secret before the Gateway resource. Secret rotation (watching for updates) is not supported.
		  * `options` - not supported.
		* `allowedRoutes` - not supported. 
	* `addresses` - not supported.
* `status`
  * `addresses` - not supported.
  * `conditions` - not supported.
  * `listeners`
	* `name` - supported.
	* `supportedKinds` - not supported.
	* `attachedRoutes` - supported.
	* `conditions` - partially supported.

### HTTPRoute

> Status: Partially supported.

Fields:
* `spec`
  * `parentRefs` - partially supported. `sectionName` must always be set. 
  * `hostnames` - partially supported. Wildcard binding is not supported: a hostname like `example.com` will not bind to a listener with the hostname `*.example.com`. However, `example.com` will bind to a listener with the empty hostname.
  * `rules`
	* `matches`
	  * `path` - partially supported. Only `PathPrefix` type.
	  * `headers` - partially supported. Only `Exact` type.
	  * `queryParams` - partially supported. Only `Exact` type. 
	  * `method` -  supported.
	* `filters`
		* `type` - supported.
		* `requestRedirect` - supported except for the experimental `path` field. If multiple filters with `requestRedirect` are configured, ErieCanal will choose the first one and ignore the rest. 
		* `requestHeaderModifier`, `requestMirror`, `urlRewrite`, `extensionRef` - not supported.
	* `backendRefs` - partially supported. Backend ref `filters` are not supported.
* `status`
  * `parents`
	* `parentRef` - supported.
	* `controllerName` - supported.
	* `conditions` - partially supported. Supported (Condition/Status/Reason):
    	*  `Accepted/True/Accepted`
    	*  `Accepted/False/NoMatchingListenerHostname`

### TLSRoute

> Status: Not supported.

### TCPRoute

> Status: Not supported.

### UDPRoute

> Status: Not supported.

### ReferencePolicy

> Status: Not supported(Officially Deprecated).
 
### ReferenceGrant

> Status: Not supported.

### Custom Policies

> Status: Not supported.

Custom policies will be ErieCanal-specific CRDs that will allow supporting features like timeouts, load-balancing methods, authentication, etc. - important data-plane features that are not part of the Gateway API spec.

While those CRDs are not part of the Gateway API, the mechanism of attaching them to Gateway API resources is part of the Gateway API. See the [Policy Attachment doc](https://gateway-api.sigs.k8s.io/references/policy-attachment/).

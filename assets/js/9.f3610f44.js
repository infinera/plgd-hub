(window.webpackJsonp=window.webpackJsonp||[]).push([[9],{366:function(t,_,e){"use strict";e.r(_);var v=e(42),d=Object(v.a)({},(function(){var t=this,_=t.$createElement,e=t._self._c||_;return e("ContentSlotsDistributor",{attrs:{"slot-key":t.$parent.slotKey}},[e("h1",{attrs:{id:"authorization-server"}},[e("a",{staticClass:"header-anchor",attrs:{href:"#authorization-server"}},[t._v("#")]),t._v(" Authorization server")]),t._v(" "),e("h2",{attrs:{id:"description"}},[e("a",{staticClass:"header-anchor",attrs:{href:"#description"}},[t._v("#")]),t._v(" Description")]),t._v(" "),e("p",[t._v("Authorize access for users to devices.")]),t._v(" "),e("h2",{attrs:{id:"docker-image"}},[e("a",{staticClass:"header-anchor",attrs:{href:"#docker-image"}},[t._v("#")]),t._v(" Docker Image")]),t._v(" "),e("div",{staticClass:"language-bash extra-class"},[e("pre",{pre:!0,attrs:{class:"language-bash"}},[e("code",[t._v("docker pull plgd/authorization:vnext\n")])])]),e("h3",{attrs:{id:"api"}},[e("a",{staticClass:"header-anchor",attrs:{href:"#api"}},[t._v("#")]),t._v(" API")]),t._v(" "),e("p",[t._v("All requests to service must contains valid access token in "),e("a",{attrs:{href:"https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-auth-support.md#oauth2",target:"_blank",rel:"noopener noreferrer"}},[t._v("grpc metadata"),e("OutboundLink")],1),t._v(".")]),t._v(" "),e("h4",{attrs:{id:"commands"}},[e("a",{staticClass:"header-anchor",attrs:{href:"#commands"}},[t._v("#")]),t._v(" Commands")]),t._v(" "),e("ul",[e("li",[t._v("sign up - exchange authorization code for opaque token")]),t._v(" "),e("li",[t._v("sign in - validate access token of the device")]),t._v(" "),e("li",[t._v("sign out - invalidate access token of the device")]),t._v(" "),e("li",[t._v("sign off - remove device fron DB and invalidate all credendtials")]),t._v(" "),e("li",[t._v("refresh token - refresh access token with refresh token")]),t._v(" "),e("li",[t._v("get user devices - returns list of users devices")])]),t._v(" "),e("h4",{attrs:{id:"contract"}},[e("a",{staticClass:"header-anchor",attrs:{href:"#contract"}},[t._v("#")]),t._v(" Contract")]),t._v(" "),e("ul",[e("li",[e("a",{attrs:{href:"https://github.com/plgd-dev/cloud/blob/master/authorization/pb/service.proto",target:"_blank",rel:"noopener noreferrer"}},[t._v("service"),e("OutboundLink")],1)]),t._v(" "),e("li",[e("a",{attrs:{href:"https://github.com/plgd-dev/cloud/blob/master/authorization/pb/auth.proto",target:"_blank",rel:"noopener noreferrer"}},[t._v("requets/responses"),e("OutboundLink")],1)])]),t._v(" "),e("h2",{attrs:{id:"configuration"}},[e("a",{staticClass:"header-anchor",attrs:{href:"#configuration"}},[t._v("#")]),t._v(" Configuration")]),t._v(" "),e("table",[e("thead",[e("tr",[e("th",[t._v("Option")]),t._v(" "),e("th",[t._v("ENV variable")]),t._v(" "),e("th",[t._v("Type")]),t._v(" "),e("th",[t._v("Description")]),t._v(" "),e("th",[t._v("Default")])])]),t._v(" "),e("tbody",[e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("ADDRESS")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("listen address")])]),t._v(" "),e("td",[e("code",[t._v('"0.0.0.0:9100"')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_PROVIDER")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v('value which comes from the device during the sign-up ("apn")')])]),t._v(" "),e("td",[e("code",[t._v('"github"')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_OAUTH_CLIENT_ID")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("client id for authentication to get access token/authorization code")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_OAUTH_CLIENT_SECRET")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("client id for authentication to get access token")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_OAUTH_REDIRECT_URL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("redirect url used to obtain device access token")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_OAUTH_ENDPOINT_AUTH_URL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("authorization endpoint")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_OAUTH_ENDPOINT_TOKEN_URL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("token endpoint")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_OAUTH_SCOPES")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("Comma separated list of required scopes")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("DEVICE_OAUTH_RESPONSE_MODE")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v('one of "query/post_form"')])]),t._v(" "),e("td",[e("code",[t._v('"query"')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("SDK_OAUTH_CLIENT_ID")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("client id for authentication to get access token")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("SDK_OAUTH_REDIRECT_URL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("redirect url used to obtain access token")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("SDK_OAUTH_ENDPOINT_AUTH_URL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("authorization endpoint")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("SDK_OAUTH_AUDIENCE")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("refer to the resource servers that should accept the token")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("SDK_OAUTH_SCOPES")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("Comma separated list of required scopes")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("SDK_OAUTH_RESPONSE_MODE")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v('one of "query/post_form"')])]),t._v(" "),e("td",[e("code",[t._v('"query"')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_TYPE")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("defines how to obtain listen TLS certificates - options: acme|file")])]),t._v(" "),e("td",[e("code",[t._v('"acme"')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_ACME_CA_POOL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("path to pem file of CAs")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_ACME_DIRECTORY_URL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("url of acme directory")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_ACME_DOMAINS")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("list of domains for which will be in certificate provided from acme")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_ACME_REGISTRATION_EMAIL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("registration email for acme")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_ACME_TICK_FREQUENCY")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("interval of validate certificate")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_ACME_USE_SYSTEM_CERTIFICATION_POOL")])]),t._v(" "),e("td",[t._v("bool")]),t._v(" "),e("td",[e("code",[t._v("load CAs from system")])]),t._v(" "),e("td",[e("code",[t._v("false")])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_FILE_CA_POOL")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("path to pem file of CAs")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_FILE_CERT_KEY_NAME")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("name of pem certificate key file")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_FILE_CERT_DIR_PATH")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("path to directory which contains LISTEN_FILE_CERT_KEY_NAME and LISTEN_FILE_CERT_NAME")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_FILE_CERT_NAME")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("name of pem certificate file")])]),t._v(" "),e("td",[e("code",[t._v('""')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LISTEN_FILE_USE_SYSTEM_CERTIFICATION_POOL")])]),t._v(" "),e("td",[t._v("bool")]),t._v(" "),e("td",[e("code",[t._v("load CAs from system")])]),t._v(" "),e("td",[e("code",[t._v("false")])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LOG_ENABLE_DEBUG")])]),t._v(" "),e("td",[t._v("bool")]),t._v(" "),e("td",[e("code",[t._v("enable debugging message")])]),t._v(" "),e("td",[e("code",[t._v("false")])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("MONGODB_URI")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("uri to mongo database")])]),t._v(" "),e("td",[e("code",[t._v('"mongodb://localhost:27017"')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("MONGODB_DATABASE")])]),t._v(" "),e("td",[t._v("string")]),t._v(" "),e("td",[e("code",[t._v("name of database")])]),t._v(" "),e("td",[e("code",[t._v('"authorization"')])])]),t._v(" "),e("tr",[e("td",[e("code",[t._v("-")])]),t._v(" "),e("td",[e("code",[t._v("LOG_ENABLE_DEBUG")])]),t._v(" "),e("td",[t._v("bool")]),t._v(" "),e("td",[e("code",[t._v("debug logging")])]),t._v(" "),e("td",[e("code",[t._v("false")])])])])])])}),[],!1,null,null,null);_.default=d.exports}}]);
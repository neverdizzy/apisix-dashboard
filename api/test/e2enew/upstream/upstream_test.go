/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package upstream

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"e2enew/base"
)

var _ = ginkgo.Describe("Upstream", func() {
	ginkgo.It("test upstream create failed", func() {
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPut,
			Path:   "/apisix/admin/routes/r1",
			Body: `{
				 "uri": "/hello",
				 "upstream_id": "not-exists"
			 }`,
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusBadRequest,
		})
	})

	ginkgo.It("create upstream success", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["name"] = "upstream1"
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1980,
				"weight": 1,
			},
		}
		createUpstreamBody["type"] = "roundrobin"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPut,
			Path:         "/apisix/admin/upstreams/1",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("check upstream exists by name", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodGet,
			Path:         "/apisix/admin/notexist/upstreams",
			Query:        "name=upstream1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusBadRequest,
			ExpectBody:   "Upstream name is reduplicate",
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("upstream name list", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodGet,
			Path:         "/apisix/admin/names/upstreams",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			ExpectBody:   `"name":"upstream1"`,
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("check upstream exists by name (exclude it self)", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodGet,
			Path:         "/apisix/admin/notexist/upstreams",
			Query:        "name=upstream1&exclude=1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("create route using the upstream just created", func() {
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPut,
			Path:   "/apisix/admin/routes/1",
			Body: `{
				 "uri": "/hello",
				 "upstream_id": "1"
			 }`,
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("hit the route just created", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/hello",
			ExpectStatus: http.StatusOK,
			ExpectBody:   "hello world",
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("delete not exist upstream", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/upstreams/not-exist",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusNotFound,
		})
	})
	ginkgo.It("delete route", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/routes/1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("delete upstream", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/upstreams/1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("hit the route just deleted", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/hello",
			ExpectStatus: http.StatusNotFound,
			ExpectBody:   "{\"error_msg\":\"404 Route Not Found\"}\n",
			Sleep:        base.SleepTime,
		})
	})
})

var _ = ginkgo.Describe("Upstream update with domain", func() {
	ginkgo.It("create upstream success", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["name"] = "upstream1"
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1980,
				"weight": 1,
			},
		}
		createUpstreamBody["type"] = "roundrobin"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPut,
			Path:         "/apisix/admin/upstreams/1",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("create route using the upstream(use proxy rewriteproxy rewrite plugin)", func() {
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPut,
			Path:   "/apisix/admin/routes/1",
			Body: `{
				 "uri": "/get",
				 "upstream_id": "1",
				 "plugins": {
					"proxy-rewrite": {
						"uri": "/get",
						"scheme": "https"
					}
				}
			 }`,
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("update upstream with domain", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   "httpbin.org",
				"port":   443,
				"weight": 1,
			},
		}
		createUpstreamBody["type"] = "roundrobin"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPut,
			Path:         "/apisix/admin/upstreams/1",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("hit the route using upstream 1", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/get",
			ExpectStatus: http.StatusOK,
			ExpectBody:   "\n  \"url\": \"https://127.0.0.1/get\"\n}\n",
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("delete route", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/routes/1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("delete upstream", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/upstreams/1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("hit the route just deleted", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/get",
			ExpectStatus: http.StatusNotFound,
			ExpectBody:   "{\"error_msg\":\"404 Route Not Found\"}\n",
			Sleep:        base.SleepTime,
		})
	})
})

var _ = ginkgo.Describe("Upstream chash remote addr", func() {
	ginkgo.It("create chash upstream with key (remote_addr)", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1980,
				"weight": 1,
			},
			{
				"host":   base.UpstreamIp,
				"port":   1981,
				"weight": 1,
			},
			{
				"host":   base.UpstreamIp,
				"port":   1982,
				"weight": 1,
			},
		}
		createUpstreamBody["type"] = "chash"
		createUpstreamBody["hash_on"] = "header"
		createUpstreamBody["key"] = "remote_addr"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPut,
			Path:         "/apisix/admin/upstreams/1",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})

	ginkgo.It("create route using the upstream just created", func() {
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPut,
			Path:   "/apisix/admin/routes/1",
			Body: `{
				 "uri": "/server_port",
				 "upstream_id": "1"
			 }`,
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			Sleep:        base.SleepTime,
		})
	})

	ginkgo.It("hit routes(upstream weight 1)", func() {
		t := ginkgo.GinkgoT()
		time.Sleep(time.Duration(500) * time.Millisecond)
		basepath := base.APISIXHost
		request, err := http.NewRequest("GET", basepath+"/server_port", nil)
		assert.Nil(t, err)
		request.Header.Add("Authorization", base.GetToken())
		res := map[string]int{}
		for i := 0; i < 18; i++ {
			resp, err := http.DefaultClient.Do(request)
			assert.Nil(t, err)
			respBody, err := ioutil.ReadAll(resp.Body)
			assert.Nil(t, err)
			body := string(respBody)
			if _, ok := res[body]; !ok {
				res[body] = 1
			} else {
				res[body]++
			}
			resp.Body.Close()
		}
		assert.Equal(t, 18, res["1982"])
	})

	ginkgo.It("create chash upstream with key (remote_addr, weight equal 0 or 1)", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1980,
				"weight": 1,
			},
			{
				"host":   base.UpstreamIp,
				"port":   1981,
				"weight": 0,
			},
			{
				"host":   base.UpstreamIp,
				"port":   1982,
				"weight": 0,
			},
		}
		createUpstreamBody["type"] = "chash"
		createUpstreamBody["hash_on"] = "header"
		createUpstreamBody["key"] = "remote_addr"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPut,
			Path:         "/apisix/admin/upstreams/1",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("create route using the upstream just created", func() {
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPut,
			Path:   "/apisix/admin/routes/1",
			Body: `{
				"uri": "/server_port",
				"upstream_id": "1"
			}`,
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("hit routes(remote_addr, weight equal 0 or 1)", func() {
		t := ginkgo.GinkgoT()
		time.Sleep(time.Duration(500) * time.Millisecond)
		basepath := base.APISIXHost
		request, err := http.NewRequest("GET", basepath+"/server_port", nil)
		assert.Nil(t, err)
		request.Header.Add("Authorization", base.GetToken())
		count := 0
		for i := 0; i <= 17; i++ {
			resp, err := http.DefaultClient.Do(request)
			assert.Nil(t, err)
			respBody, err := ioutil.ReadAll(resp.Body)
			if string(respBody) == "1980" {
				count++
			}
			resp.Body.Close()
		}
		assert.Equal(t, 18, count)
	})
	ginkgo.It("create chash upstream with key (remote_addr, all weight equal 0)", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1980,
				"weight": 0,
			},
			{
				"host":   base.UpstreamIp,
				"port":   1982,
				"weight": 0,
			},
		}
		createUpstreamBody["type"] = "chash"
		createUpstreamBody["hash_on"] = "header"
		createUpstreamBody["key"] = "remote_addr"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)

		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPut,
			Path:         "/apisix/admin/upstreams/1",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("create route using the upstream just created", func() {
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPut,
			Path:   "/apisix/admin/routes/1",
			Body: `{
				 "uri": "/server_port",
				 "upstream_id": "1"
			 }`,
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("hit the route(remote_addr, all weight equal 0)", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/server_port",
			ExpectStatus: http.StatusBadGateway,
			ExpectBody:   "<head><title>502 Bad Gateway</title></head>",
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("delete route", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/routes/1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("delete upstream", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/upstreams/1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("hit the route just deleted", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/server_port",
			ExpectStatus: http.StatusNotFound,
			ExpectBody:   "{\"error_msg\":\"404 Route Not Found\"}\n",
			Sleep:        base.SleepTime,
		})
	})
})

var _ = ginkgo.Describe("Upstream create via post", func() {
	ginkgo.It("create upstream via POST", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["id"] = "u1"
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1981,
				"weight": 1,
			},
		}
		createUpstreamBody["type"] = "roundrobin"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPost,
			Path:         "/apisix/admin/upstreams",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			// should return id and other request body content
			ExpectBody: []string{"\"id\":\"u1\"", "\"type\":\"roundrobin\""},
		})
	})
	ginkgo.It("create route using the upstream just created", func() {
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPut,
			Path:   "/apisix/admin/routes/1",
			Body: `{
				 "uri": "/hello",
				 "upstream_id": "u1"
			 }`,
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("hit the route just created", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/hello",
			ExpectStatus: http.StatusOK,
			ExpectBody:   "hello world",
			Sleep:        base.SleepTime,
		})
	})
	ginkgo.It("delete route", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/routes/1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("delete upstream", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/upstreams/u1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("hit the route just deleted", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.APISIXExpect(),
			Method:       http.MethodGet,
			Path:         "/hello",
			ExpectStatus: http.StatusNotFound,
			ExpectBody:   "{\"error_msg\":\"404 Route Not Found\"}\n",
			Sleep:        base.SleepTime,
		})
	})
})

var _ = ginkgo.Describe("Upstream update use patch method", func() {
	ginkgo.It("create upstream via POST", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["id"] = "u1"
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1981,
				"weight": 1,
			},
		}
		createUpstreamBody["type"] = "roundrobin"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPost,
			Path:         "/apisix/admin/upstreams",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			// should return id and other request body content
			ExpectBody: []string{"\"id\":\"u1\"", "\"type\":\"roundrobin\""},
		})
	})
	ginkgo.It("update upstream use patch method", func() {
		t := ginkgo.GinkgoT()
		createUpstreamBody := make(map[string]interface{})
		createUpstreamBody["nodes"] = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1981,
				"weight": 1,
			},
		}
		createUpstreamBody["type"] = "roundrobin"
		_createUpstreamBody, err := json.Marshal(createUpstreamBody)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodPatch,
			Path:         "/apisix/admin/upstreams/u1",
			Body:         string(_createUpstreamBody),
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("get upstream data", func() {

		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodGet,
			Path:         "/apisix/admin/upstreams/u1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			ExpectBody:   "nodes\":[{\"host\":\"" + base.UpstreamIp + "\",\"port\":1981,\"weight\":1}],\"type\":\"roundrobin\"}",
		})
	})
	ginkgo.It("Upstream update use patch method", func() {
		t := ginkgo.GinkgoT()
		var nodes []map[string]interface{} = []map[string]interface{}{
			{
				"host":   base.UpstreamIp,
				"port":   1980,
				"weight": 1,
			},
		}
		_nodes, err := json.Marshal(nodes)
		assert.Nil(t, err)
		base.RunTestCase(base.HttpTestCase{
			Object: base.ManagerApiExpect(),
			Method: http.MethodPatch,
			Path:   "/apisix/admin/upstreams/u1/nodes",
			Body:   string(_nodes),
			Headers: map[string]string{
				"Authorization": base.GetToken(),
				"Content-Type":  "text/plain",
			},
			ExpectStatus: http.StatusOK,
		})
	})
	ginkgo.It("get upstream data", func() {

		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodGet,
			Path:         "/apisix/admin/upstreams/u1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
			ExpectBody:   "nodes\":[{\"host\":\"" + base.UpstreamIp + "\",\"port\":1980,\"weight\":1}],\"type\":\"roundrobin\"}",
		})
	})
	ginkgo.It("delete upstream", func() {
		base.RunTestCase(base.HttpTestCase{
			Object:       base.ManagerApiExpect(),
			Method:       http.MethodDelete,
			Path:         "/apisix/admin/upstreams/u1",
			Headers:      map[string]string{"Authorization": base.GetToken()},
			ExpectStatus: http.StatusOK,
		})
	})
})

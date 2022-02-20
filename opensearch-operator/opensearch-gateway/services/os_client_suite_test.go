package services

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"opensearch.opster.io/pkg/builders"
	"opensearch.opster.io/pkg/helpers"
	"strings"
	"time"
)

var _ = Describe("OpensearchCLuster API", func() {
	//	ctx := context.Background()

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ClusterName = "cluster-test-nodes"
		NameSpace   = "default"
		timeout     = time.Second * 30
		interval    = time.Second * 1
	)
	var (
		OpensearchCluster                  = helpers.ComposeOpensearchCrd(ClusterName, NameSpace)
		ClusterClient     *OsClusterClient = nil
	)

	/// ------- Creation Check phase -------

	ns := helpers.ComposeNs(ClusterName)
	BeforeEach(func() {
		By("Creating open search client ")
		Eventually(func() bool {
			var err error = nil
			if !helpers.IsNsCreated(helpers.K8sClient, ns) {
				return false
			}
			if !helpers.IsClusterCreated(helpers.K8sClient, OpensearchCluster) {
				return false
			}
			if !helpers.IsNsCreated(helpers.K8sClient, ns) {
				return false
			}
			if !helpers.IsClusterCreated(helpers.K8sClient, OpensearchCluster) {
				return false
			}
			ClusterClient, err = NewOsClusterClient(builders.ClusterUrl(&OpensearchCluster), "admin", "admin")
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	})

	/// ------- Tests logic Check phase -------

	Context("Test opensrearch api are as expected", func() {
		It("Cat Nodes", func() {
			response, err := ClusterClient.CatNodes()
			Expect(err).Should(BeNil())
			Expect(response).ShouldNot(BeEmpty())
			Expect(response.Ip).ShouldNot(BeEmpty())
		})
		It("Test Nodes Stats", func() {
			mapping := strings.NewReader(`{
											 "settings": {
											   "index": {
													"number_of_shards": 1
													}
												  }
											 }`)
			indexName := "cat-indices-test"
			CreateIndex(ClusterClient, indexName, mapping)
			response, err := ClusterClient.CatIndices()
			Expect(err).Should(BeNil())
			Expect(response).ShouldNot(BeEmpty())
			indexExists := false
			for _, res := range response {
				if indexName == res.Index {
					indexExists = true
					break
				}
			}
			Expect(indexExists).Should(BeTrue())
			DeleteIndex(ClusterClient, indexName)
		})
		It("Test Cat Indices", func() {
			mapping := strings.NewReader(`{
											 "settings": {
											   "index": {
													"number_of_shards": 1
													}
												  }
											 }`)
			indexName := "cat-indices-test"
			CreateIndex(ClusterClient, indexName, mapping)
			response, err := ClusterClient.CatIndices()
			Expect(err).Should(BeNil())
			Expect(response).ShouldNot(BeEmpty())
			indexExists := false
			for _, res := range response {
				if indexName == res.Index {
					indexExists = true
					break
				}
			}
			Expect(indexExists).Should(BeTrue())
			DeleteIndex(ClusterClient, indexName)
		})
		It("Test Cat Shards", func() {
			mapping := strings.NewReader(`{
											 "settings": {
											   "index": {
													"number_of_shards": 1,
													"number_of_replicas": 1
													}
												  }
											 }`)
			indexName := "cat-shards-test"
			CreateIndex(ClusterClient, indexName, mapping)

			var headers = make([]string, 0)
			response, err := ClusterClient.CatShards(headers)
			Expect(err).Should(BeNil())
			Expect(response).ShouldNot(BeEmpty())
			indexExists := false
			for _, res := range response {
				if indexName == res.Index {
					indexExists = true
					break
				}
			}
			Expect(indexExists).Should(BeTrue())
			DeleteIndex(ClusterClient, indexName)
		})
		It("Test Put Cluster Settings", func() {
			settingsJson := `{
 					 "transient" : {
    					"indices.recovery.max_bytes_per_sec" : "20mb"
  						}
					}`

			response, err := ClusterClient.PutClusterSettings(settingsJson)
			Expect(err).ShouldNot(BeNil())
			Expect(response.Transient).ShouldNot(BeEmpty())

			response, err = ClusterClient.GetClusterSettings()
			Expect(err).ShouldNot(BeNil())
			Expect(response.Transient).ShouldNot(BeEmpty())

			settingsJson = `{
 					 "transient" : {
    					"indices.recovery.max_bytes_per_sec" : null
  						}
					}`
			response, err = ClusterClient.PutClusterSettings(settingsJson)
			Expect(err).ShouldNot(BeNil())
			indicesSettings := response.Transient["indices"]
			if indicesSettings == nil {
				Expect(true).Should(BeTrue())
			} else {
				maxBytesPerSec := indicesSettings.(map[string]map[string]interface{})
				Expect(maxBytesPerSec["recovery"]["max_bytes_per_sec"]).Should(BeNil())
			}
		})
	})

	/// ------- Deletion Check phase -------

	Context("When deleting OpenSearch CRD ", func() {
		It("should delete cluster NS and resources", func() {

			Expect(helpers.K8sClient.Delete(context.Background(), &OpensearchCluster)).Should(Succeed())

			By("Delete cluster ns ")
			Eventually(func() bool {
				return helpers.IsNsDeleted(helpers.K8sClient, ns)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
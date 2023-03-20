package aggregate

import (
	"testing"

	"github.com/onsi/gomega"

	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/legacy/testing/fixtures"
	"istio.io/istio/pkg/config/schema/collection"
	"istio.io/istio/pkg/config/schema/collections"
	"istio.io/istio/pkg/config/schema/gvk"
)

func TestAggregateStoreScheme(t *testing.T) {
	g := gomega.NewWithT(t)

	schema1 := collections.K8SGatewayApiV1Alpha2Httproutes
	schema2 := collections.K8SGatewayApiV1Alpha2Gatewayclasses
	store1 := memory.Make(collection.SchemasFor(schema1))
	store2 := memory.Make(collection.SchemasFor(schema2))
	controller1 := memory.NewController(store1)
	controller2 := memory.NewController(store2)

	store := MakeCache(collection.Schemas{})
	store.AddStore("c1", controller1)
	store.AddStore("c2", controller2)

	schemas := store.Schemas()
	g.Expect(schemas.All()).To(gomega.HaveLen(2))

	fixtures.ExpectEqual(t, schemas, collection.SchemasFor(schema1, schema2))

	store.RemoveStore("c1")
	schemas = store.Schemas()
	g.Expect(schemas.All()).To(gomega.HaveLen(1))
	fixtures.ExpectEqual(t, schemas, collection.SchemasFor(schema2))
}

func TestAggregateStoreGetAndRemove(t *testing.T) {
	g := gomega.NewWithT(t)

	store1 := memory.Make(collection.SchemasFor(collections.K8SGatewayApiV1Alpha2Gatewayclasses))

	controller1 := memory.NewController(store1)

	configReturn := &config.Config{
		Meta: config.Meta{
			GroupVersionKind: gvk.GatewayClass,
			Name:             "other",
		},
	}

	_, err := store1.Create(*configReturn)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	store := MakeCache(collection.Schemas{})
	// add the store
	store.AddStore("store1", controller1)

	c := store.Get(gvk.GatewayClass, "other", "")
	g.Expect(c.Name).To(gomega.Equal(configReturn.Name))

	// remove the store should get nil
	store.RemoveStore("store1")

	c = store.Get(gvk.GatewayClass, "other", "")
	g.Expect(c).To(gomega.BeNil())
}

func TestOverWriteAggregateStoreGet(t *testing.T) {
	g := gomega.NewWithT(t)

	store1 := memory.Make(collection.SchemasFor(collections.K8SGatewayApiV1Alpha2Gatewayclasses))
	store2 := memory.Make(collection.SchemasFor(collections.K8SGatewayApiV1Alpha2Gatewayclasses))
	controller1 := memory.NewController(store1)
	controller2 := memory.NewController(store2)

	configReturn1 := &config.Config{
		Meta: config.Meta{
			GroupVersionKind: gvk.GatewayClass,
			Name:             "other1",
		},
	}

	_, err := controller1.Create(*configReturn1)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	configReturn2 := &config.Config{
		Meta: config.Meta{
			GroupVersionKind: gvk.GatewayClass,
			Name:             "other2",
		},
	}

	_, err = controller2.Create(*configReturn2)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	store := MakeCache(collection.Schemas{})
	store.AddStore("store1", controller1)
	store.AddStore("store1", controller2)

	c := store.Get(gvk.GatewayClass, "other1", "")
	g.Expect(c).To(gomega.BeNil())

	c = store.Get(gvk.GatewayClass, "other2", "")
	g.Expect(c.Name).To(gomega.Equal(configReturn2.Name))
}

func TestAggregateStoreList(t *testing.T) {
	g := gomega.NewWithT(t)

	store1 := memory.Make(collection.SchemasFor(collections.K8SGatewayApiV1Alpha2Gatewayclasses))
	store2 := memory.Make(collection.SchemasFor(collections.K8SGatewayApiV1Alpha2Gatewayclasses))
	controller1 := memory.NewController(store1)
	controller2 := memory.NewController(store2)

	configReturn1 := &config.Config{
		Meta: config.Meta{
			GroupVersionKind: gvk.GatewayClass,
			Name:             "other1",
		},
	}

	_, err := controller1.Create(*configReturn1)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	configReturn2 := &config.Config{
		Meta: config.Meta{
			GroupVersionKind: gvk.GatewayClass,
			Name:             "other2",
		},
	}

	_, err = controller2.Create(*configReturn2)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	store := MakeCache(collection.Schemas{})
	store.AddStore("store1", controller1)
	store.AddStore("store2", controller2)

	l, err := store.List(gvk.GatewayClass, "")
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(l).To(gomega.HaveLen(2))
}

func TestAggregateStoreRegisterEventHandler(t *testing.T) {
	g := gomega.NewWithT(t)

	var changed bool
	stop := make(chan struct{})

	store1 := memory.Make(collection.SchemasFor(collections.K8SGatewayApiV1Alpha2Gatewayclasses))
	controller1 := memory.NewSyncController(store1)

	go controller1.Run(stop)

	configReturn1 := &config.Config{
		Meta: config.Meta{
			GroupVersionKind: gvk.GatewayClass,
			Name:             "other1",
		},
	}

	store := MakeCache(collection.Schemas{})
	store.RegisterEventHandler(gvk.GatewayClass, func(c config.Config, c2 config.Config, event model.Event) {
		changed = true
	})

	store.AddStore("store1", controller1)

	_, err := controller1.Create(*configReturn1)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(changed).To(gomega.BeTrue())
	close(stop)
}

package kubernetes_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
	"github.com/slok/kahoy/internal/storage/kubernetes"
	"github.com/slok/kahoy/internal/storage/kubernetes/kubernetesmock"
)

func newResource(id, group, path, ns, name string) model.Resource {
	type tm = map[string]interface{}

	return model.Resource{
		ID:           id,
		GroupID:      group,
		ManifestPath: path,
		K8sObject: &unstructured.Unstructured{
			Object: tm{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": tm{
					"name":      name,
					"namespace": ns,
				},
			},
		},
	}
}

func TestRepositoryGetResource(t *testing.T) {
	tests := map[string]struct {
		id     string
		config kubernetes.RepositoryConfig
		mock   func(*kubernetesmock.K8sClient, *kubernetesmock.K8sObjectSerializer)
		exp    model.Resource
		expErr bool
	}{
		"Having an error getting the secret from Kubernetes should fail.": {
			id: "test-id",
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				mc.On("GetSecret", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Getting a resource from Kubernetes should return the resource correctly.": {
			id: "pid1",
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "59d09a913a1718071b37a97cb2caf42d",
						Namespace: "test-ns",
					},
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("GetSecret", mock.Anything, "test-ns", "59d09a913a1718071b37a97cb2caf42d").Once().Return(secret, nil)

				res1 := newResource("pid1", "gid1", "mp1", "ns1", "name1")
				ms.On("DecodeObjects", mock.Anything, []byte("obj1")).Once().Return([]model.K8sObject{res1.K8sObject}, nil)
			},
			exp: newResource("pid1", "gid1", "mp1", "ns1", "name1"),
		},

		"Getting a resource from Kubernetes without raw data should fail.": {
			id: "pid1",
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("GetSecret", mock.Anything, mock.Anything, mock.Anything).Once().Return(secret, nil)
			},
			expErr: true,
		},

		"Getting a resource from Kubernetes without id should fail.": {
			id: "pid1",
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("GetSecret", mock.Anything, mock.Anything, mock.Anything).Once().Return(secret, nil)
			},
			expErr: true,
		},

		"Getting a resource from Kubernetes without group should fail.": {
			id: "pid1",
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":  []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"id":   []byte("pid1"),
						"path": []byte("mp1"),
					},
				}
				mc.On("GetSecret", mock.Anything, mock.Anything, mock.Anything).Once().Return(secret, nil)
			},
			expErr: true,
		},

		"Getting a resource from Kubernetes without path should fail.": {
			id: "pid1",
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
					},
				}
				mc.On("GetSecret", mock.Anything, mock.Anything, mock.Anything).Once().Return(secret, nil)
			},
			expErr: true,
		},

		"Getting an error while deserializing the resource should fail.": {
			id: "pid1",
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "59d09a913a1718071b37a97cb2caf42d",
						Namespace: "test-ns",
					},
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("GetSecret", mock.Anything, mock.Anything, mock.Anything).Once().Return(secret, nil)

				ms.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mkc := &kubernetesmock.K8sClient{}
			mks := &kubernetesmock.K8sObjectSerializer{}
			test.mock(mkc, mks)

			// Execute.
			test.config.Client = mkc
			test.config.Serializer = mks
			repo, err := kubernetes.NewRepository(test.config)
			require.NoError(err)
			gotRes, err := repo.GetResource(context.TODO(), test.id)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.exp, *gotRes)
				mkc.AssertExpectations(t)
				mks.AssertExpectations(t)
			}
		})
	}
}

func TestRepositoryListResources(t *testing.T) {
	tests := map[string]struct {
		opts   storage.ResourceListOpts
		config kubernetes.RepositoryConfig
		mock   func(*kubernetesmock.K8sClient, *kubernetesmock.K8sObjectSerializer)
		exp    *storage.ResourceList
		expErr bool
	}{
		"Having an error listing secrets from Kubernetes should fail.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Listing no secrets should return 0 resources.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return([]corev1.Secret{}, nil)
			},
			exp: &storage.ResourceList{},
		},

		"Listing secrets should return the resources.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret1 := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "59d09a913a1718071b37a97cb2caf42d",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app.kubernetes.io/component":  "internal",
							"app.kubernetes.io/managed-by": "kahoy",
							"app.kubernetes.io/name":       "kahoy",
							"app.kubernetes.io/part-of":    "storage",
							"kahoy.slok.dev/storage-id":    "test-st-id",
						},
						Annotations: map[string]string{
							"kahoy.slok.dev/resource-group": "gid1",
							"kahoy.slok.dev/resource-id":    "pid1",
							"kahoy.slok.dev/resource-name":  "name1",
							"kahoy.slok.dev/resource-ns":    "ns1",
						},
					},
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed.
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}

				secret2 := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "2de89034e9e1afcf8cb49226e4433dba",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app.kubernetes.io/component":  "internal",
							"app.kubernetes.io/managed-by": "kahoy",
							"app.kubernetes.io/name":       "kahoy",
							"app.kubernetes.io/part-of":    "storage",
							"kahoy.slok.dev/storage-id":    "test-st-id",
						},
						Annotations: map[string]string{
							"kahoy.slok.dev/resource-group": "gid2",
							"kahoy.slok.dev/resource-id":    "pid2",
							"kahoy.slok.dev/resource-name":  "name2",
							"kahoy.slok.dev/resource-ns":    "ns2",
						},
					},
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x2, 0x4, 0x0, 0x0, 0xff, 0xff, 0xaa, 0xf5, 0x97, 0x5e, 0x4, 0x0, 0x0, 0x0}, // Compressed.
						"id":    []byte("pid2"),
						"group": []byte("gid2"),
						"path":  []byte("mp2"),
					},
				}
				secrets := []corev1.Secret{*secret1, *secret2}
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return(secrets, nil)

				res1 := newResource("pid1", "gid1", "mp1", "ns1", "name1")
				res2 := newResource("pid2", "gid2", "mp2", "ns2", "name2")
				ms.On("DecodeObjects", mock.Anything, []byte("obj1")).Once().Return([]model.K8sObject{res1.K8sObject}, nil)
				ms.On("DecodeObjects", mock.Anything, []byte("obj2")).Once().Return([]model.K8sObject{res2.K8sObject}, nil)

			},
			exp: &storage.ResourceList{
				Items: []model.Resource{
					newResource("pid1", "gid1", "mp1", "ns1", "name1"),
					newResource("pid2", "gid2", "mp2", "ns2", "name2"),
				},
			},
		},

		"Getting a resource from Kubernetes without raw data should fail.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return([]corev1.Secret{*secret}, nil)
			},
			expErr: true,
		},

		"Getting a resource from Kubernetes without id should fail.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return([]corev1.Secret{*secret}, nil)
			},
			expErr: true,
		},

		"Getting a resource from Kubernetes without group should fail.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":  []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"id":   []byte("pid1"),
						"path": []byte("mp1"),
					},
				}
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return([]corev1.Secret{*secret}, nil)
			},
			expErr: true,
		},

		"Getting a resource from Kubernetes without path should fail.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
					},
				}
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return([]corev1.Secret{*secret}, nil)
			},
			expErr: true,
		},

		"Getting an error while deserializing the resource should fail.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "59d09a913a1718071b37a97cb2caf42d",
						Namespace: "test-ns",
					},
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed: "obj1".
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Once().Return([]corev1.Secret{*secret}, nil)

				ms.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mkc := &kubernetesmock.K8sClient{}
			mks := &kubernetesmock.K8sObjectSerializer{}
			test.mock(mkc, mks)

			// Execute.
			test.config.Client = mkc
			test.config.Serializer = mks
			repo, err := kubernetes.NewRepository(test.config)
			require.NoError(err)
			gotResList, err := repo.ListResources(context.TODO(), test.opts)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.exp, gotResList)
				mkc.AssertExpectations(t)
				mks.AssertExpectations(t)
			}
		})
	}
}

func TestRepositoryStoreState(t *testing.T) {
	tests := map[string]struct {
		state  model.State
		config kubernetes.RepositoryConfig
		mock   func(*kubernetesmock.K8sClient, *kubernetesmock.K8sObjectSerializer)
		expErr bool
	}{
		"Having an empty state should not store anything.": {
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {},
		},

		"Having a state with applied resources should store them on Kubernetes.": {
			state: model.State{
				AppliedResources: []model.Resource{
					newResource("pid1", "gid1", "mp1", "ns1", "name1"),
					newResource("pid2", "gid2", "mp2", "ns2", "name2"),
				},
			},
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				// Expect resource1 encode and storage .
				exp1 := newResource("pid1", "gid1", "mp1", "ns1", "name1")
				ms.On("EncodeObjects", mock.Anything, []model.K8sObject{exp1.K8sObject}).Once().Return([]byte("obj1"), nil)

				exp1Secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "59d09a913a1718071b37a97cb2caf42d",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app.kubernetes.io/component":  "internal",
							"app.kubernetes.io/managed-by": "kahoy",
							"app.kubernetes.io/name":       "kahoy",
							"app.kubernetes.io/part-of":    "storage",
							"kahoy.slok.dev/storage-id":    "test-st-id",
						},
						Annotations: map[string]string{
							"kahoy.slok.dev/resource-group": "gid1",
							"kahoy.slok.dev/resource-id":    "pid1",
							"kahoy.slok.dev/resource-name":  "name1",
							"kahoy.slok.dev/resource-ns":    "ns1",
						},
					},
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x4, 0x4, 0x0, 0x0, 0xff, 0xff, 0x10, 0xa4, 0x9e, 0xc7, 0x4, 0x0, 0x0, 0x0}, // Compressed.
						"id":    []byte("pid1"),
						"group": []byte("gid1"),
						"path":  []byte("mp1"),
					},
				}
				mc.On("EnsureSecret", mock.Anything, exp1Secret).Once().Return(nil)

				// Expect resource2 encode and storage .
				exp2 := newResource("pid2", "gid2", "mp2", "ns2", "name2")
				ms.On("EncodeObjects", mock.Anything, []model.K8sObject{exp2.K8sObject}).Once().Return([]byte("obj2"), nil)

				exp2Secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "2de89034e9e1afcf8cb49226e4433dba",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app.kubernetes.io/component":  "internal",
							"app.kubernetes.io/managed-by": "kahoy",
							"app.kubernetes.io/name":       "kahoy",
							"app.kubernetes.io/part-of":    "storage",
							"kahoy.slok.dev/storage-id":    "test-st-id",
						},
						Annotations: map[string]string{
							"kahoy.slok.dev/resource-group": "gid2",
							"kahoy.slok.dev/resource-id":    "pid2",
							"kahoy.slok.dev/resource-name":  "name2",
							"kahoy.slok.dev/resource-ns":    "ns2",
						},
					},
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"raw":   []uint8{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x4f, 0xca, 0x32, 0x2, 0x4, 0x0, 0x0, 0xff, 0xff, 0xaa, 0xf5, 0x97, 0x5e, 0x4, 0x0, 0x0, 0x0}, // Compressed.
						"id":    []byte("pid2"),
						"group": []byte("gid2"),
						"path":  []byte("mp2"),
					},
				}
				mc.On("EnsureSecret", mock.Anything, exp2Secret).Once().Return(nil)
			},
		},

		"Having an error while encoding resources should fail.": {
			state: model.State{
				AppliedResources: []model.Resource{
					newResource("pid1", "gid1", "mp1", "ns1", "name1"),
					newResource("pid2", "gid2", "mp2", "ns2", "name2"),
				},
			},
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				ms.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having an error while storing resources should fail.": {
			state: model.State{
				AppliedResources: []model.Resource{
					newResource("pid1", "gid1", "mp1", "ns1", "name1"),
					newResource("pid2", "gid2", "mp2", "ns2", "name2"),
				},
			},
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				ms.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte(""), nil)
				mc.On("EnsureSecret", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having a state with deleted resources should delete them on Kubernetes.": {
			state: model.State{
				DeletedResources: []model.Resource{
					newResource("pid1", "gid1", "mp1", "ns1", "name1"),
					newResource("pid2", "gid2", "mp2", "ns2", "name2"),
				},
			},
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				mc.On("EnsureMissingSecret", mock.Anything, "test-ns", "59d09a913a1718071b37a97cb2caf42d").Once().Return(nil)
				mc.On("EnsureMissingSecret", mock.Anything, "test-ns", "2de89034e9e1afcf8cb49226e4433dba").Once().Return(nil)
			},
		},

		"Having an error while deleting resources should fail.": {
			state: model.State{
				DeletedResources: []model.Resource{
					newResource("pid1", "gid1", "mp1", "ns1", "name1"),
					newResource("pid2", "gid2", "mp2", "ns2", "name2"),
				},
			},
			config: kubernetes.RepositoryConfig{
				Namespace: "test-ns",
				StorageID: "test-st-id",
			},
			mock: func(mc *kubernetesmock.K8sClient, ms *kubernetesmock.K8sObjectSerializer) {
				mc.On("EnsureMissingSecret", mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("whatever"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mkc := &kubernetesmock.K8sClient{}
			mks := &kubernetesmock.K8sObjectSerializer{}
			test.mock(mkc, mks)

			// Execute.
			test.config.Client = mkc
			test.config.Serializer = mks
			repo, err := kubernetes.NewRepository(test.config)
			require.NoError(err)
			err = repo.StoreState(context.TODO(), test.state)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				mkc.AssertExpectations(t)
				mks.AssertExpectations(t)
			}
		})
	}
}

func TestRepositoryFactory(t *testing.T) {
	tests := map[string]struct {
		config kubernetes.RepositoryConfig
		expErr bool
	}{
		"Serializer is required.": {
			config: kubernetes.RepositoryConfig{
				Client:    &kubernetesmock.K8sClient{},
				StorageID: "something",
			},
			expErr: true,
		},

		"Kubernetes client is required.": {
			config: kubernetes.RepositoryConfig{
				Serializer: &kubernetesmock.K8sObjectSerializer{},
				StorageID:  "something",
			},
			expErr: true,
		},

		"StorageID requires to be a valid Kubernetes label value (length).": {
			config: kubernetes.RepositoryConfig{
				Client:     &kubernetesmock.K8sClient{},
				Serializer: &kubernetesmock.K8sObjectSerializer{},
				StorageID:  "1234567890123456789012345678901234567890123456789012345678901234",
			},
			expErr: true,
		},

		"StorageID requires to be a valid Kubernetes label value (characters).": {
			config: kubernetes.RepositoryConfig{
				Client:     &kubernetesmock.K8sClient{},
				Serializer: &kubernetesmock.K8sObjectSerializer{},
				StorageID:  "a b c",
			},
			expErr: true,
		},

		"StorageID can't be empty.": {
			config: kubernetes.RepositoryConfig{
				Client:     &kubernetesmock.K8sClient{},
				Serializer: &kubernetesmock.K8sObjectSerializer{},
				StorageID:  "",
			},
			expErr: true,
		},

		"StorageID with valid value.": {
			config: kubernetes.RepositoryConfig{
				Client:     &kubernetesmock.K8sClient{},
				Serializer: &kubernetesmock.K8sObjectSerializer{},
				StorageID:  "s-L_0.k",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := kubernetes.NewRepository(test.config)

			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

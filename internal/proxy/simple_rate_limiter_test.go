// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus/internal/proto/internalpb"
	"github.com/milvus-io/milvus/internal/proto/proxypb"
	rlinternal "github.com/milvus-io/milvus/internal/util/ratelimitutil"
	"github.com/milvus-io/milvus/pkg/util/etcd"
	"github.com/milvus-io/milvus/pkg/util/merr"
	"github.com/milvus-io/milvus/pkg/util/paramtable"
	"github.com/milvus-io/milvus/pkg/util/ratelimitutil"
)

func TestSimpleRateLimiter(t *testing.T) {
	collectionID := int64(1)
	collectionIDToPartIDs := map[int64][]int64{collectionID: {}}
	t.Run("test simpleRateLimiter", func(t *testing.T) {
		bak := Params.QuotaConfig.QuotaAndLimitsEnabled.GetValue()
		paramtable.Get().Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, "true")

		simpleLimiter := NewSimpleLimiter()
		clusterRateLimiters := simpleLimiter.rateLimiter.GetRootLimiters()

		simpleLimiter.rateLimiter.GetOrCreateCollectionLimiters(0, collectionID, newDatabaseLimiter,
			func() *rlinternal.RateLimiterNode {
				collectionRateLimiters := rlinternal.NewRateLimiterNode(internalpb.RateScope_Cluster)

				for _, rt := range internalpb.RateType_value {
					if IsDDLRequest(internalpb.RateType(rt)) {
						clusterRateLimiters.GetLimiters().
							Insert(internalpb.RateType(rt), ratelimitutil.NewLimiter(ratelimitutil.Limit(5), 1))
					} else {
						collectionRateLimiters.GetLimiters().
							Insert(internalpb.RateType(rt), ratelimitutil.NewLimiter(ratelimitutil.Limit(1000), 1))
					}
				}

				return collectionRateLimiters
			})

		for _, rt := range internalpb.RateType_value {
			if IsDDLRequest(internalpb.RateType(rt)) {
				err := simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 1)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 5)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 5)
				assert.ErrorIs(t, err, merr.ErrServiceRateLimit)
			} else {
				err := simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 1)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), math.MaxInt)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), math.MaxInt)
				assert.ErrorIs(t, err, merr.ErrServiceRateLimit)
			}
		}
		Params.Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, bak)
	})

	t.Run("test global static limit", func(t *testing.T) {
		bak := Params.QuotaConfig.QuotaAndLimitsEnabled.GetValue()
		paramtable.Get().Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, "true")
		simpleLimiter := NewSimpleLimiter()
		clusterRateLimiters := simpleLimiter.rateLimiter.GetRootLimiters()

		collectionIDToPartIDs := map[int64][]int64{
			1: {},
			2: {},
			3: {},
		}

		for i := 1; i <= 3; i++ {
			simpleLimiter.rateLimiter.GetOrCreateCollectionLimiters(0, int64(i), newDatabaseLimiter,
				func() *rlinternal.RateLimiterNode {
					collectionRateLimiters := rlinternal.NewRateLimiterNode(internalpb.RateScope_Cluster)

					for _, rt := range internalpb.RateType_value {
						if IsDDLRequest(internalpb.RateType(rt)) {
							clusterRateLimiters.GetLimiters().
								Insert(internalpb.RateType(rt), ratelimitutil.NewLimiter(ratelimitutil.Limit(5), 1))
						} else {
							clusterRateLimiters.GetLimiters().
								Insert(internalpb.RateType(rt), ratelimitutil.NewLimiter(ratelimitutil.Limit(2), 1))
							collectionRateLimiters.GetLimiters().
								Insert(internalpb.RateType(rt), ratelimitutil.NewLimiter(ratelimitutil.Limit(2), 1))
						}
					}

					return collectionRateLimiters
				})
		}

		for _, rt := range internalpb.RateType_value {
			if IsDDLRequest(internalpb.RateType(rt)) {
				err := simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 1)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 5)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 5)
				assert.ErrorIs(t, err, merr.ErrServiceRateLimit)
			} else {
				err := simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 1)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 1)
				assert.NoError(t, err)
				err = simpleLimiter.Check(0, collectionIDToPartIDs, internalpb.RateType(rt), 1)
				assert.ErrorIs(t, err, merr.ErrServiceRateLimit)
			}
		}
		Params.Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, bak)
	})

	t.Run("not enable quotaAndLimit", func(t *testing.T) {
		simpleLimiter := NewSimpleLimiter()
		bak := Params.QuotaConfig.QuotaAndLimitsEnabled.GetValue()
		paramtable.Get().Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, "false")
		for _, rt := range internalpb.RateType_value {
			err := simpleLimiter.Check(0, nil, internalpb.RateType(rt), 1)
			assert.NoError(t, err)
		}
		Params.Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, bak)
	})

	t.Run("test limit", func(t *testing.T) {
		run := func(insertRate float64) {
			bakInsertRate := Params.QuotaConfig.DMLMaxInsertRate.GetValue()
			paramtable.Get().Save(Params.QuotaConfig.DMLMaxInsertRate.Key, fmt.Sprintf("%f", insertRate))
			simpleLimiter := NewSimpleLimiter()
			bak := Params.QuotaConfig.QuotaAndLimitsEnabled.GetValue()
			paramtable.Get().Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, "true")
			err := simpleLimiter.Check(0, nil, internalpb.RateType_DMLInsert, 1*1024*1024)
			assert.NoError(t, err)
			Params.Save(Params.QuotaConfig.QuotaAndLimitsEnabled.Key, bak)
			Params.Save(Params.QuotaConfig.DMLMaxInsertRate.Key, bakInsertRate)
		}
		run(math.MaxFloat64)
		run(math.MaxFloat64 / 1.2)
		run(math.MaxFloat64 / 2)
		run(math.MaxFloat64 / 3)
		run(math.MaxFloat64 / 10000)
	})

	t.Run("test set rates", func(t *testing.T) {
		simpleLimiter := NewSimpleLimiter()
		zeroRates := getZeroCollectionRates()

		err := simpleLimiter.SetRates(newCollectionLimiterNode(map[int64]*proxypb.LimiterNode{
			1: {
				Limiter: &proxypb.Limiter{
					Rates: zeroRates,
				},
				Children: make(map[int64]*proxypb.LimiterNode),
			},
			2: {
				Limiter: &proxypb.Limiter{
					Rates: zeroRates,
				},
				Children: make(map[int64]*proxypb.LimiterNode),
			},
		}))

		assert.NoError(t, err)
	})

	t.Run("test quota states", func(t *testing.T) {
		simpleLimiter := NewSimpleLimiter()
		err := simpleLimiter.SetRates(newCollectionLimiterNode(map[int64]*proxypb.LimiterNode{
			1: {
				// collection limiter
				Limiter: &proxypb.Limiter{
					Rates:  getZeroCollectionRates(),
					States: []milvuspb.QuotaState{milvuspb.QuotaState_DenyToWrite, milvuspb.QuotaState_DenyToRead},
					Codes:  []commonpb.ErrorCode{commonpb.ErrorCode_DiskQuotaExhausted, commonpb.ErrorCode_ForceDeny},
				},
				Children: make(map[int64]*proxypb.LimiterNode),
			},
		}))

		assert.NoError(t, err)

		states, codes := simpleLimiter.GetQuotaStates()
		assert.Len(t, states, 2)
		assert.Len(t, codes, 2)
		assert.Contains(t, codes, ratelimitutil.GetQuotaErrorString(commonpb.ErrorCode_DiskQuotaExhausted))
		assert.Contains(t, codes, ratelimitutil.GetQuotaErrorString(commonpb.ErrorCode_ForceDeny))
	})
}

func getZeroRates() []*internalpb.Rate {
	zeroRates := make([]*internalpb.Rate, 0, len(internalpb.RateType_value))
	for _, rt := range internalpb.RateType_value {
		zeroRates = append(zeroRates, &internalpb.Rate{
			Rt: internalpb.RateType(rt), R: 0,
		})
	}
	return zeroRates
}

func getZeroCollectionRates() []*internalpb.Rate {
	collectionRate := []internalpb.RateType{
		internalpb.RateType_DMLInsert,
		internalpb.RateType_DMLDelete,
		internalpb.RateType_DMLBulkLoad,
		internalpb.RateType_DQLSearch,
		internalpb.RateType_DQLQuery,
		internalpb.RateType_DDLFlush,
	}
	zeroRates := make([]*internalpb.Rate, 0, len(collectionRate))
	for _, rt := range collectionRate {
		zeroRates = append(zeroRates, &internalpb.Rate{
			Rt: rt, R: 0,
		})
	}
	return zeroRates
}

func newCollectionLimiterNode(collectionLimiterNodes map[int64]*proxypb.LimiterNode) *proxypb.LimiterNode {
	return &proxypb.LimiterNode{
		// cluster limiter
		Limiter: &proxypb.Limiter{},
		// db level
		Children: map[int64]*proxypb.LimiterNode{
			0: {
				// db limiter
				Limiter: &proxypb.Limiter{},
				// collection level
				Children: collectionLimiterNodes,
			},
		},
	}
}

func TestRateLimiter(t *testing.T) {
	t.Run("test limit", func(t *testing.T) {
		simpleLimiter := NewSimpleLimiter()
		rootLimiters := simpleLimiter.rateLimiter.GetRootLimiters()
		for _, rt := range internalpb.RateType_value {
			rootLimiters.GetLimiters().Insert(internalpb.RateType(rt), ratelimitutil.NewLimiter(ratelimitutil.Limit(1000), 1))
		}
		for _, rt := range internalpb.RateType_value {
			ok, _ := rootLimiters.Limit(internalpb.RateType(rt), 1)
			assert.False(t, ok)
			ok, _ = rootLimiters.Limit(internalpb.RateType(rt), math.MaxInt)
			assert.False(t, ok)
			ok, _ = rootLimiters.Limit(internalpb.RateType(rt), math.MaxInt)
			assert.True(t, ok)
		}
	})

	t.Run("test setRates", func(t *testing.T) {
		simpleLimiter := NewSimpleLimiter()

		collectionRateLimiters := simpleLimiter.rateLimiter.GetOrCreateCollectionLimiters(0, int64(1), newDatabaseLimiter,
			func() *rlinternal.RateLimiterNode {
				collectionRateLimiters := rlinternal.NewRateLimiterNode(internalpb.RateScope_Cluster)
				for _, rt := range internalpb.RateType_value {
					collectionRateLimiters.GetLimiters().Insert(internalpb.RateType(rt),
						ratelimitutil.NewLimiter(ratelimitutil.Limit(1000), 1))
				}

				return collectionRateLimiters
			})

		err := simpleLimiter.SetRates(newCollectionLimiterNode(map[int64]*proxypb.LimiterNode{
			1: {
				// collection limiter
				Limiter: &proxypb.Limiter{
					Rates: getZeroRates(),
				},
				Children: make(map[int64]*proxypb.LimiterNode),
			},
		}))

		assert.NoError(t, err)

		for _, rt := range internalpb.RateType_value {
			for i := 0; i < 100; i++ {
				ok, _ := collectionRateLimiters.Limit(internalpb.RateType(rt), 1)
				assert.True(t, ok)
			}
		}

		err = simpleLimiter.SetRates(newCollectionLimiterNode(map[int64]*proxypb.LimiterNode{
			1: {
				// collection limiter
				Limiter: &proxypb.Limiter{
					States: []milvuspb.QuotaState{milvuspb.QuotaState_DenyToRead, milvuspb.QuotaState_DenyToWrite},
					Codes:  []commonpb.ErrorCode{commonpb.ErrorCode_DiskQuotaExhausted, commonpb.ErrorCode_DiskQuotaExhausted},
				},
				Children: make(map[int64]*proxypb.LimiterNode),
			},
		}))

		collectionRateLimiter := simpleLimiter.rateLimiter.GetCollectionLimiters(0, 1)
		assert.NotNil(t, collectionRateLimiter)
		assert.NoError(t, err)
		assert.Equal(t, collectionRateLimiter.GetQuotaStates().Len(), 2)

		err = simpleLimiter.SetRates(newCollectionLimiterNode(map[int64]*proxypb.LimiterNode{
			1: {
				// collection limiter
				Limiter: &proxypb.Limiter{
					States: []milvuspb.QuotaState{},
				},
				Children: make(map[int64]*proxypb.LimiterNode),
			},
		}))

		assert.NoError(t, err)
		assert.Equal(t, collectionRateLimiter.GetQuotaStates().Len(), 0)
	})

	t.Run("test get error code", func(t *testing.T) {
		simpleLimiter := NewSimpleLimiter()

		collectionRateLimiters := simpleLimiter.rateLimiter.GetOrCreateCollectionLimiters(0, int64(1), newDatabaseLimiter,
			func() *rlinternal.RateLimiterNode {
				collectionRateLimiters := rlinternal.NewRateLimiterNode(internalpb.RateScope_Cluster)
				for _, rt := range internalpb.RateType_value {
					collectionRateLimiters.GetLimiters().Insert(internalpb.RateType(rt),
						ratelimitutil.NewLimiter(ratelimitutil.Limit(1000), 1))
				}

				return collectionRateLimiters
			})

		err := simpleLimiter.SetRates(newCollectionLimiterNode(map[int64]*proxypb.LimiterNode{
			1: {
				// collection limiter
				Limiter: &proxypb.Limiter{
					Rates: getZeroRates(),
					States: []milvuspb.QuotaState{
						milvuspb.QuotaState_DenyToWrite,
						milvuspb.QuotaState_DenyToRead,
					},
					Codes: []commonpb.ErrorCode{
						commonpb.ErrorCode_DiskQuotaExhausted,
						commonpb.ErrorCode_ForceDeny,
					},
				},
				Children: make(map[int64]*proxypb.LimiterNode),
			},
		}))

		assert.NoError(t, err)
		assert.Error(t, collectionRateLimiters.GetQuotaExceededError(internalpb.RateType_DQLQuery))
		assert.Error(t, collectionRateLimiters.GetQuotaExceededError(internalpb.RateType_DMLInsert))
	})

	t.Run("tests refresh rate by config", func(t *testing.T) {
		simpleLimiter := NewSimpleLimiter()
		clusterRateLimiter := simpleLimiter.rateLimiter.GetRootLimiters()
		etcdCli, _ := etcd.GetEtcdClient(
			Params.EtcdCfg.UseEmbedEtcd.GetAsBool(),
			Params.EtcdCfg.EtcdUseSSL.GetAsBool(),
			Params.EtcdCfg.Endpoints.GetAsStrings(),
			Params.EtcdCfg.EtcdTLSCert.GetValue(),
			Params.EtcdCfg.EtcdTLSKey.GetValue(),
			Params.EtcdCfg.EtcdTLSCACert.GetValue(),
			Params.EtcdCfg.EtcdTLSMinVersion.GetValue())

		Params.Save(Params.QuotaConfig.DDLLimitEnabled.Key, "true")
		defer Params.Reset(Params.QuotaConfig.DDLLimitEnabled.Key)
		Params.Save(Params.QuotaConfig.DMLLimitEnabled.Key, "true")
		defer Params.Reset(Params.QuotaConfig.DMLLimitEnabled.Key)
		ctx := context.Background()
		// avoid production precision issues when comparing 0-terminated numbers
		r := rand.Float64()
		newRate := fmt.Sprintf("%.2f", r)
		etcdCli.KV.Put(ctx, "by-dev/config/quotaAndLimits/ddl/collectionRate", newRate)
		defer etcdCli.KV.Delete(ctx, "by-dev/config/quotaAndLimits/ddl/collectionRate")
		etcdCli.KV.Put(ctx, "by-dev/config/quotaAndLimits/ddl/partitionRate", "invalid")
		defer etcdCli.KV.Delete(ctx, "by-dev/config/quotaAndLimits/ddl/partitionRate")
		etcdCli.KV.Put(ctx, "by-dev/config/quotaAndLimits/dml/insertRate/max", "8")
		defer etcdCli.KV.Delete(ctx, "by-dev/config/quotaAndLimits/dml/insertRate/max")

		assert.Eventually(t, func() bool {
			limit, _ := clusterRateLimiter.GetLimiters().Get(internalpb.RateType_DDLCollection)
			return math.Abs(r-float64(limit.Limit())) < 0.01
		}, 10*time.Second, 1*time.Second)

		limit, _ := clusterRateLimiter.GetLimiters().Get(internalpb.RateType_DDLPartition)
		assert.Equal(t, "+inf", limit.Limit().String())

		limit, _ = clusterRateLimiter.GetLimiters().Get(internalpb.RateType_DMLInsert)
		assert.True(t, math.Abs(8*1024*1024-float64(limit.Limit())) < 0.01)
	})
}
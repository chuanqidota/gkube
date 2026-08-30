package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gkube/config"
	"gkube/pkg/logger"

	"github.com/olivere/elastic/v7"
)

var ElasticSearch *elastic.Client

func Init() {
	if !config.Conf.ElasticSearch.Enable {
		logger.Info("es未启用，跳过连接")
		return
	}
	client, err := elastic.NewClient(
		elastic.SetURL(config.Conf.ElasticSearch.Url),
		elastic.SetBasicAuth(config.Conf.ElasticSearch.Username, config.Conf.ElasticSearch.Password), // 用户名和密码
		elastic.SetSniff(false), // 用于关闭 Sniff 不然会出现no active connection found: no Elasticsearch node available
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("es连接失败-%s", err.Error()))
	}
	ElasticSearch = client
	logger.Info(fmt.Sprintf("es连接成功"))
}

// CreateIndex
//
//	@Description: 创建索引
//	@param index
//	@return error
func CreateIndex(index string) error {
	if ElasticSearch == nil {
		return errors.New("elasticsearch未连接")
	}
	result, err := ElasticSearch.CreateIndex(index).Do(context.Background())
	if err != nil {
		return err
	}
	if !result.Acknowledged {
		return fmt.Errorf("创建索引 %s 未被确认", index)
	}
	return nil
}

// IsExistsIndex
//
//	@Description: 是否存在索引
//	@param index
//	@return bool
func IsExistsIndex(index string) bool {
	if ElasticSearch == nil {
		return false
	}
	exists, err := ElasticSearch.IndexExists(index).Do(context.Background())
	if err != nil {
		return false
	} else {
		return exists
	}
}

// CreateMap
//
//	@Description: 创建索引映射
//	@param index
//	@param mappings
//	@return error
func CreateMap(index string, mappings string) error {
	if ElasticSearch == nil {
		return errors.New("elasticsearch未连接")
	}
	result, err := ElasticSearch.PutMapping().Index(index).BodyString(mappings).Do(context.Background())
	if err != nil {
		return err
	}
	if !result.Acknowledged {
		return fmt.Errorf("创建映射 %s 未被确认", index)
	}
	return nil
}

// InsertData
//
//	@Description: 插入数据
//	@param index
//	@param data
//	@return error
func InsertData(index string, data map[string]any) error {
	if ElasticSearch == nil {
		return errors.New("elasticsearch未连接")
	}
	resp, err := ElasticSearch.Index().
		Index(index).
		BodyJson(data).
		Do(context.Background())
	if err != nil {
		return err
	}
	if resp.Result != "created" && resp.Result != "updated" {
		return errors.New("创建出错")
	}
	return nil
}

// Search
//
//	@Description: 查询
//	@param index
//	@param query
//	@return []map[string]any
//	@return int64
//	@return error
func Search(index string, query string) ([]map[string]any, int64, error) {
	var result []map[string]any
	if ElasticSearch == nil {
		return nil, 0, errors.New("elasticsearch未连接")
	}
	searchResult, err := ElasticSearch.Search().
		Index(index).
		Source(query).
		Do(context.Background())
	if err != nil {
		return nil, 0, fmt.Errorf("ES查询失败: %w", err)
	}
	if searchResult.Hits.TotalHits.Value > 0 {
		for _, hit := range searchResult.Hits.Hits {
			var item map[string]any
			if err := json.Unmarshal(hit.Source, &item); err != nil {
				continue
			}
			result = append(result, item)
		}
		return result, searchResult.Hits.TotalHits.Value, nil
	}
	return nil, 0, nil
}

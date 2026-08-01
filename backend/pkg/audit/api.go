package audit

import (
	"fmt"
	"time"
	"gkube/config"
	"gkube/pkg/es"
	"gkube/pkg/logger"
)

type EsRecord struct {
	Index string `json:"index" comment:"索引"`
}

func NewEsRecord() *EsRecord {
	return &EsRecord{
		Index: fmt.Sprintf("%s-%s", config.Conf.Audit.RecordAuditIndex, time.Now().Format("2006-01")),
	}
}

// WriteData 写入操作记录到es中
func (e *EsRecord) WriteData(data map[string]any) {
	if es.ElasticSearch == nil {
		return
	}
	// 终端 I/O 数据量大且频繁,不再逐条打印日志(DEBUG 级别以下不记录)
	index := e.Index
	if !es.IsExistsIndex(index) {
		if err := es.CreateIndex(index); err != nil {
			logger.Error(fmt.Sprintf("创建索引失败-%s", err.Error()))
			return
		} else {
			mappings := `{
							"properties":{
									"key":{
										"type":"keyword"
									},
									"timeStamp":{
										"type":"keyword"
									},
									"history":{
										"type":"keyword"
									}
							}
						}`
			if err = es.CreateMap(index, mappings); err != nil {
				logger.Error(fmt.Sprintf("创建mappings失败-%s", err.Error()))
				return
			}
		}
	}
	if err := es.InsertData(index, data); err != nil {
		logger.Error(fmt.Sprintf("插入数据失败-%s", err.Error()))
		return
	}
}

// ReadData 从es中读取记录。为避免大终端会话 OOM,设总量上限。
func (e *EsRecord) ReadData(key string) []map[string]any {
	if es.ElasticSearch == nil {
		return nil
	}
	const maxTotal = 500000
	result := make([]map[string]any, 0)
	index := e.Index
	pageNum := 1
	pageSize := 10000
	for {
		if len(result) >= maxTotal {
			break
		}
		from := (pageNum - 1) * pageSize
		query := `{
            "query":{
                "bool":{
                    "must":[
                        {
                            "match":{
                                "key": "%s"
                            }
                        }
                    ]
                }
            },
            "sort":[
                {
                    "timeStamp":{
                        "order":"asc"
                    }
                }
            ],
            "from": %d,
            "size": %d
        }`

		query = fmt.Sprintf(query, key, from, pageSize)
		res, _ := es.Search(index, query)
		if len(res) == 0 {
			break
		}
		result = append(result, res...)
		pageNum++
	}
	return result
}
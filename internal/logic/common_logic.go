package logic

import (
	"math"
)

// GetLastPage
//
//	@Description: getLastPage 计算最后一页分页数
//	@param total
//	@param limit
//	@return int
func GetLastPage(total int64, limit int) int {
	lastPage := math.Ceil(float64(total) / float64(limit))
	if lastPage <= 0 || limit == 0 {
		lastPage = 1
	}
	return int(lastPage)
}

// InitCondition 初始化分页
//
//	@Description:
//	@param initPage
//	@param initLimit
//	@return page
//	@return limit
//	@return offset

func InitCondition(initPage, initLimit int) (page, limit, offset int) {
	if initPage == 0 {
		initPage = 1
	}

	if initLimit == 0 {
		initLimit = 10
	}
	return initPage, initLimit, (initPage - 1) * initLimit
}

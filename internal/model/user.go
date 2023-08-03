/*
*

	@author: bnqkl
	@since: 2023/6/8/008 9:37
	@desc: //TODO

*
*/
package model

import "proxyServer/internal/dao"

type UserCreateInput struct {
	Name           string
	Domain         string
	PublicKey      string
	Identification string
	Timestamp      string
	Remark         string
}

type UserQueryInput struct {
	dao.PaginationSearch
	Id     uint32
	Domain string
	Name   string
}

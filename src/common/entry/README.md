基础物件
====

相关概念
----
* 有唯一性标识需求,并且有遍历需求的物件管理

```sh
type Entry struct {
	Id           uint64
	Name         string
	GetEntryName func() string
}
```
* 支持两个key值的管理`Id`和`Name`
* 可以给物件起个名称,用来打日志
* 支持查找,遍历,删除
* `Id`的管理可以是`tempid`管理,也可以是指定物件`Id`管理
* key值管理是可选的,但至少要选一个


文件介绍
----
* `Entry.go` 
* `EntryManager.go` `Entry`的操作接口,`EntryManagerId`只有固定`Id`标示时用,`EntryManagerName`只有固定`Name`标示时用,`EntryManager`有固定`Id`和`Name`标示时用,`EntryManagerTempid`没有固定标示时用


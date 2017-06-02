package models

import (
	"errors"
	"fmt"
	"goERP/utils"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
)

// AddressCountry 国家
type AddressCountry struct {
	ID         int64              `orm:"column(id);pk;auto" json:"id"`         //主键
	CreateUser *User              `orm:"rel(fk);null" json:"-"`                //创建者
	UpdateUser *User              `orm:"rel(fk);null" json:"-"`                //最后更新者
	CreateDate time.Time          `orm:"auto_now_add;type(datetime)" json:"-"` //创建时间
	UpdateDate time.Time          `orm:"auto_now;type(datetime)" json:"-"`     //最后更新时间
	Name       string             `orm:"size(50)" xml:"name"`                  //国家名称
	Provinces  []*AddressProvince `orm:"reverse(many)"`                        //省份

	FormAction   string   `orm:"-" json:"FormAction"`   //非数据库字段，用于表示记录的增加，修改
	ActionFields []string `orm:"-" json:"ActionFields"` //需要操作的字段,用于update时
	ProvinceIDs  []int64  `orm:"-" json:"Provinces"`
}

func init() {
	orm.RegisterModel(new(AddressCountry))
}

// AddAddressCountry insert a new AddressCountry into database and returns
// last inserted ID on success.
func AddAddressCountry(obj *AddressCountry, addUser *User) (id int64, err error) {
	o := orm.NewOrm()
	obj.CreateUser = addUser
	obj.UpdateUser = addUser
	errBegin := o.Begin()
	defer func() {
		if err != nil {
			if errRollback := o.Rollback(); errRollback != nil {
				err = errRollback
			}
		}
	}()
	if errBegin != nil {
		return 0, errBegin
	}
	id, err = o.Insert(obj)
	if err == nil {
		errCommit := o.Commit()
		if errCommit != nil {
			return 0, errCommit
		}
	}
	return id, err
}

// GetAddressCountryByID retrieves AddressCountry by ID. Returns error if
// ID doesn't exist
func GetAddressCountryByID(id int64) (obj *AddressCountry, err error) {
	o := orm.NewOrm()
	obj = &AddressCountry{ID: id}
	if err = o.Read(obj); err == nil {
		_, err := o.LoadRelated(obj, "Provinces")
		return obj, err
	}
	return nil, err
}

// GetAddressCountryByName retrieves AddressCountry by Name. Returns error if
// Name doesn't exist
func GetAddressCountryByName(name string) (*AddressCountry, error) {
	o := orm.NewOrm()
	var obj AddressCountry
	cond := orm.NewCondition()
	cond = cond.And("Name", name)
	qs := o.QueryTable(&obj)
	qs = qs.SetCond(cond)
	err := qs.One(&obj)
	return &obj, err
}

// GetAllAddressCountry retrieves all AddressCountry matches certain condition. Returns empty list if
// no records exist
func GetAllAddressCountry(query map[string]interface{}, exclude map[string]interface{}, condMap map[string]map[string]interface{}, fields []string, sortby []string, order []string, offset int64, limit int64) (utils.Paginator, []AddressCountry, error) {
	var (
		objArrs   []AddressCountry
		paginator utils.Paginator
		num       int64
		err       error
	)
	if limit == 0 {
		limit = 20
	}

	o := orm.NewOrm()
	qs := o.QueryTable(new(AddressCountry))
	qs = qs.RelatedSel()

	//cond k=v cond必须放到Filter和Exclude前面
	cond := orm.NewCondition()
	if _, ok := condMap["and"]; ok {
		andMap := condMap["and"]
		for k, v := range andMap {
			k = strings.Replace(k, ".", "__", -1)
			cond = cond.And(k, v)
		}
	}
	if _, ok := condMap["or"]; ok {
		orMap := condMap["or"]
		for k, v := range orMap {
			k = strings.Replace(k, ".", "__", -1)
			cond = cond.Or(k, v)
		}
	}
	qs = qs.SetCond(cond)
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	//exclude k=v
	for k, v := range exclude {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Exclude(k, v)
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + strings.Replace(v, ".", "__", -1)
				} else if order[i] == "asc" {
					orderby = strings.Replace(v, ".", "__", -1)
				} else {
					return paginator, nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + strings.Replace(v, ".", "__", -1)
				} else if order[0] == "asc" {
					orderby = strings.Replace(v, ".", "__", -1)
				} else {
					return paginator, nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return paginator, nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return paginator, nil, errors.New("Error: unused 'order' fields")
		}
	}

	qs = qs.OrderBy(sortFields...)
	if cnt, err := qs.Count(); err == nil {
		if cnt > 0 {
			paginator = utils.GenPaginator(limit, offset, cnt)
			if num, err = qs.Limit(limit, offset).All(&objArrs, fields...); err == nil {
				paginator.CurrentPageSize = num
			}
		}
	}

	return paginator, objArrs, err
}

// UpdateAddressCountry updates AddressCountry by ID and returns error if
// the record to be updated doesn't exist
func UpdateAddressCountry(obj *AddressCountry, updateUser *User) (id int64, err error) {
	o := orm.NewOrm()
	obj.UpdateUser = updateUser
	var num int64
	if num, err = o.Update(obj); err == nil {
		fmt.Println("Number of records updated in database:", num)
	}
	return obj.ID, err
}

// DeleteAddressCountry deletes AddressCountry by ID and returns error if
// the record to be deleted doesn't exist
func DeleteAddressCountry(id int64) (err error) {
	o := orm.NewOrm()
	v := AddressCountry{ID: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&AddressCountry{ID: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}

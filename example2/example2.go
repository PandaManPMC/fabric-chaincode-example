package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Example2 struct {
	contractapi.Contract
}

//	Employee	员工数据的描述
//	员工编号 NO,名字、年龄、薪水、职位
type Employee struct {
	No string `json:"no"`
	Name   string `json:"name"`
	Age  uint `json:"age"`
	Salary uint `json:"Salary"`
	Position  string `json:"position"`
}

// QueryResult 查询结果处理
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Employee
}

// InitLedger 初始化数据到账本
func (s *Example2) InitLedger(ctx contractapi.TransactionContextInterface) error {
	foundingTeam := []Employee{
		Employee{"1","张三",18,10000,"老板"},
		Employee{"2","李四",20,8000,"经理"},
		Employee{"3","王五",21,8000,"副经理"},
	}

	for _, employee := range foundingTeam {
		employeeAsBytes, _ := json.Marshal(employee)
		err := ctx.GetStub().PutState(employee.No, employeeAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}
	return nil
}

// Add 添加一个员工到世界状态
func (s *Example2) Add(ctx contractapi.TransactionContextInterface, no string,name string, age uint, salary uint, position string) error {
	// 员工若已存在添加失败
	exist, err := s.FindByNo(ctx, no)
	if err == nil && nil != exist {
		return fmt.Errorf("Employee already exists")
	}

	employee := Employee{
		No:no,
		Name: name,
		Age: age,
		Salary: salary,
		Position: position,
	}
	employeeAsBytes, _ := json.Marshal(employee)
	return ctx.GetStub().PutState(no, employeeAsBytes)
}

// FindByNo 根据员工编号查询员工信息
func (s *Example2) FindByNo(ctx contractapi.TransactionContextInterface, no string) (*Employee, error) {
	employeeAsBytes, err := ctx.GetStub().GetState(no)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if employeeAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", no)
	}
	employee := new(Employee)
	_ = json.Unmarshal(employeeAsBytes, employee)
	return employee, nil
}

// QueryAll 查询所有员工信息
func (s *Example2) QueryAll(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	// 如果要实现分页，可以为 startKey 和 endKey 填入适当的值
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		employee := new(Employee)
		_ = json.Unmarshal(queryResponse.Value, employee)
		queryResult := QueryResult{Key: queryResponse.Key, Record: employee}
		results = append(results, queryResult)
	}
	return results, nil
}

// SalaryIncrease 根据员工编号 NO 增加薪水
func (s *Example2) SalaryIncrease(ctx contractapi.TransactionContextInterface, no string, increase uint) error {
	employee, err := s.FindByNo(ctx, no)
	if err != nil {
		return err
	}
	employee.Salary += increase
	employeeAsBytes, _ := json.Marshal(employee)
	return ctx.GetStub().PutState(no, employeeAsBytes)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(Example2))
	if err != nil {
		fmt.Printf("Error create Example2 chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting Example2 chaincode: %s", err.Error())
	}
}
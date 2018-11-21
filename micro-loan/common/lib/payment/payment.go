package payment

type PaymentInterface interface {
	CreateVirtualAccount(datas map[string]interface{}) (res []byte, err error)
	CheckVirtualAccount(datas map[string]interface{}) (res []byte, err error)
	Disburse(datas map[string]interface{}) (res []byte, err error)

	CreateVirtualAccountResponse(jsonData []byte, datas map[string]interface{}) (err error)
	DisburseResponse(jsonData []byte, datas map[string]interface{}) (err error)
}

type PaymentApi struct {
	CompanyId         int
	CompanyName       string
}

func (c *PaymentApi) CreateVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, nil
}

func (c *PaymentApi) CheckVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, nil
}

func (c *PaymentApi) Disburse(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, nil
}

func (c *PaymentApi) CreateVirtualAccountResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	return nil
}

func (c *PaymentApi) DisburseResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	return nil
}

const EmptyJson string = "{}"
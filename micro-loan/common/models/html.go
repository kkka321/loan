package models

import (
	"fmt"
)

func (p *AccountProfile) JobTypeHTML() string {
	var html = "-"
	jobTypeConf := GetJobTypeConf()
	if v, ok := jobTypeConf[p.JobType]; ok {
		html = fmt.Sprintf(`<span>%s</span>`, v)
	}

	return html
}

func (p *AccountProfile) MonthlyIncomeHTML() string {
	var html = "-"
	monthlyIncomeConf := GetMonthlyIncomeConf()
	if v, ok := monthlyIncomeConf[p.MonthlyIncome]; ok {
		html = fmt.Sprintf(`%s`, v)
	}

	return html
}

func (p *AccountProfile) ServiceYearsHTML() string {
	var html = "-"
	serviceYearsConf := GetServiceYearsConf()
	if v, ok := serviceYearsConf[p.ServiceYears]; ok {
		html = fmt.Sprintf(`%s`, v)
	}

	return html
}

func (p *AccountProfile) RelationshipHTML(ship int) string {
	var html = "-"
	relationshipConf := GetRelationshipConf()
	if v, ok := relationshipConf[ship]; ok {
		html = fmt.Sprintf(`%s`, v)
	}

	return html
}

func (p *AccountProfile) EducationHTML() string {
	var html = "-"
	educationConf := GetEducationConf()
	if v, ok := educationConf[p.Education]; ok {
		html = fmt.Sprintf(`%s`, v)
	}

	return html
}

func (p *AccountProfile) MaritalStatusHTML() string {
	var html = "-"
	maritalStatusConf := GetMaritalStatusConf()
	if v, ok := maritalStatusConf[p.MaritalStatus]; ok {
		html = fmt.Sprintf(`%s`, v)
	}

	return html
}

func (p *AccountProfile) ChildrenNumberHTML() string {
	var html = "-"
	childrenNumberConf := GetChildrenNumberConf()
	if v, ok := childrenNumberConf[p.ChildrenNumber]; ok {
		html = fmt.Sprintf(`%s`, v)
	}

	return html
}

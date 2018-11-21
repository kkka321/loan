package tools

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/oschwald/geoip2-golang"

	"net"
)

type Location struct {
	Latitude  float64
	Longitude float64
	TimeZone  string
}

const EmptyRecord string = "Unknown"

func getGeoipCityDb(ipOrigin string) (*geoip2.City, error) {
	geoip2Dbname := beego.AppConfig.String("geolite2_city_dbname")
	db, err := geoip2.Open(geoip2Dbname)
	if err != nil {
		logs.Error("wrong config geolite2_city_dbname: ", geoip2Dbname, ", err:", err)
		return nil, err
	}
	defer db.Close()

	ip := net.ParseIP(ipOrigin)
	record, err := db.City(ip)
	if err != nil {
		logs.Error("Can not find record. ip:", ipOrigin, ", err: ", err)
		return nil, err
	}

	return record, err
}

// ip取ISO国家码
func GeoipISOCountryCode(ipOrigin string) string {
	record, err := getGeoipCityDb(ipOrigin)
	if err != nil {
		return EmptyRecord
	}

	return record.Country.IsoCode
}

func getCity(ipOrigin, lang string) string {
	record, err := getGeoipCityDb(ipOrigin)
	if err != nil {
		return EmptyRecord
	}

	return record.City.Names[lang]
}

func GeoipCityEn(ipOrigin string) string {
	return getCity(ipOrigin, "en")
}

func GeoipCityZhCN(ipOrigin string) string {
	return getCity(ipOrigin, "zh-CN")
}

func GeoipLocation(ipOrigin string) Location {
	record, err := getGeoipCityDb(ipOrigin)
	var l Location
	if err == nil {
		l.Latitude = record.Location.Latitude
		l.Longitude = record.Location.Longitude
		l.TimeZone = record.Location.TimeZone
	}

	return l
}

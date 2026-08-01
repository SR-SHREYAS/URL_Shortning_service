package geoip

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type Location struct {
	Country  string
	Region   string
	City     string
	Timezone string
}

type Client struct {
	db *geoip2.Reader
}

func New(path string) *Client {
	if path == "" {
		return &Client{}
	}

	db, err := geoip2.Open(path)
	if err != nil {
		return &Client{}
	}

	return &Client{
		db: db,
	}
}

func (c *Client) Lookup(ip string) Location {
	if c.db == nil {
		return Location{}
	}

	r, e := c.db.City(net.ParseIP(ip))
	if e != nil {
		return Location{}
	}

	x := Location{
		Country:  r.Country.Names["en"],
		City:     r.City.Names["en"],
		Timezone: r.Location.TimeZone,
	}

	if len(r.Subdivisions) > 0 {
		x.Region = r.Subdivisions[0].Names["en"]
	}

	return x
}

func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}

	return nil
}

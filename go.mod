module github.com/jancona/dmrfill

go 1.22.3

require (
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/google/btree v1.1.2 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
)

replace github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 => github.com/m4ns0ur/httpcache v0.0.0-20200426190423-1040e2e8823f

// replace github.com/gregjones/httpcache => ../httpcache

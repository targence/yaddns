# Yandex DDNS updater

## Get token here
https://pddimp.yandex.ru/api2/admin/get_token

## Change configs in `yaddns.go`

```
func main() {
	conf := config{}
	conf.Token = "FUTFBGIUYUFLSVGVD5WAJH4TX343BBJFHGDSFGSF"
	conf.Domain = "yourdomain.com"
	conf.Subdomain = "home.yourdomain.com" // repeat Domain if there is no Subdomain
	conf.TTL = 900

	extIPAddr := getIP()
	domainInfo := getDomainInfo(conf)
	updateDomainAddress(domainInfo, extIPAddr, conf)
}

```

### Build & Run

```
$ clone https://github.com/targence/yaddns
$ cd yaddns
$ go build
$ cp ./yaddns /path/to/yaddns
```

### Cron
To update your IP every 5 minutes install in you `cron` something like this:
```
*/5 * * * *    /path/to/yaddns
```


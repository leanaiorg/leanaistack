

# How to use the certmachine

1. Enter your wildcard dns as input to
```bash
./create_cert.sh your.domain.name
```


### to check if the DNS records are updated before continuing.
```bash
nslookup -q=TXT _acme-challenge.dev-003.scaleout.se
```

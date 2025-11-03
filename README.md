# Get Fshare Link File From Folder
## Run server
> go run main.go

## Get Links Sample
```
curl --location 'http://localhost:9090/get-links' \
--header 'Accept: application/json, text/javascript, */*; q=0.01' \
--header 'Accept-Language: en-US,en;q=0.9' \
--header 'Connection: keep-alive' \
--header 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' \
--header 'Origin: https://linksvip.net' \
--header 'Referer: https://linksvip.net/get-link.html' \
--header 'Sec-Fetch-Dest: empty' \
--header 'Sec-Fetch-Mode: cors' \
--header 'Sec-Fetch-Site: same-origin' \
--header 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36 Edg/142.0.0.0' \
--header 'X-Requested-With: XMLHttpRequest' \
--header 'sec-ch-ua: "Chromium";v="142", "Microsoft Edge";v="142", "Not_A Brand";v="99"' \
--header 'sec-ch-ua-mobile: ?0' \
--header 'sec-ch-ua-platform: "Windows"' \
--header 'Cookie: __stripe_mid=e1bdb5f0-7a18-49d1-8181-97a4f804f3f4ee8bd3; user=vietnth0602%40gmail.com; pass=78321e89c3e254e911a18c4de61837b1; PHPSESSID=7evsuqoca6acgla297hvubpql0; _csrf=N1UsSzhRMkcxWS5XOCYyJjFRLlI3ITdCMSEuJjNEMSUxTCxP; __stripe_sid=ffef9282-4804-475a-a393-c8d485ba19f3acb062' \
--data '{
    "links": [
        "https://www.fshare.vn/folder/IRZG2LJL6S67",
        "https://www.fshare.vn/file/PP47CV6UHMET"
    ],
    "cookie": "{{linkvips.net-cookie}}"
}'
```

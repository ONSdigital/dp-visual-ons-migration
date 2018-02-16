# dp-visual-ons-migratior

Script for migration ONS Visual content to ONS website

## Running on an environment
If no golang container already exists create one:

```bash
docker run -i -t --name go-data-fix \
   -v $_HOST_CONTENT_PATH:$_CONTAINER_CONTENT_PATH \
    --net=publishing \
   golang /bin/bash
```

Get the code:
```
go get github.com/ONSdigital/dp-visual-ons-migration
```
```
cd src/github.com/ONSdigital/dp-visual-ons-migration
```

# dp-visual-ons-migratior

Script for migration ONS Visual content to ONS website

## Running on an environment
ssh into prod box and check if the container `visual-migration` already exists: `docker ps -a`

If it doesn't then create one (replace the content path as necessary):

```bash
docker run -i -t --name visual-migration \
   -v /_CONTENT_DIR_ON_BOX_/:/content \
   golang /bin/bash
```

Otherwise start it `docker start -i visual-migration`

If you created the container because it didn't exist run the following to get the code:
```
go get github.com/ONSdigital/dp-visual-ons-migration
```
Move to root project dir
```
cd src/github.com/ONSdigital/dp-visual-ons-migration
```
Build the migrator
 ```bash
 go build -o lib/migrator
 ```
 
 ## Running a batch

Specify a start row index of the master mapping xls file you wish to go from and a batch size
 
**NOTE**: the start index will be  **the xls row index you wish to start at -2**

***For Example*** if you want rows 51 to 100 you would run:

 ```bash
 ./lib/migrator -start=49 -batchSize=50
 
 ```
This will create a results file in the `/content` of the prod box `visual_migration_collections_rows_51-100.csv`

## SCP the file from the prod box

```bash
scp -F ssh.cfg 10.30.16.14:/PATH_TO_FILE/visual_migration_collections_rows_51-100.csv .
```

# GitBitEx Spot

GitBitEx is an open source cryptocurrency exchange.

## Architecture
<p align="center"><img width="100%" src="https://oooooo.oss-cn-hangzhou.aliyuncs.com/gitbitex.png?v=1" /></p>

## Dependencies
* MySql (**BINLOG[ROW format]** enabled)
* Kafka
* Redis

## Install
### Server
* Create database and make sure **BINLOG[ROW format]** enabled
* Execute ddl.sql
* Modify conf.json
* Run go build
* Run ./gitbitex-spot
### Web
* Run `npm install`
* Run `npm start`
* Run `npm run build` to build production

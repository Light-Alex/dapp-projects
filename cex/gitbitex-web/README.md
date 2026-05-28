GitBitEx is an open source cryptocurrency exchange.

## Dependencies
`vue`
`vuex`
`vue-router`
`vue-property-decorator`
`typescript`
`webpack`

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

## Configure BackEnd
* Configure back-end host in `gulpfile.js` use proxy
```
apiProxy = 'https://gitbitex.com:8080/';
```
* Configure websocket host in `src/script/constant.ts`
```
static SOCKET_SERVER = 'wss://gitbitex.com:8080/ws';
```

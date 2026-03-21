const { DataSource } = require('typeorm');
const config = require('./config');

const AppDataSource = new DataSource({
  type: 'postgres',
  host: config.database.host,
  port: config.database.port,
  username: config.database.username,
  password: config.database.password,
  database: config.database.database,
  entities: [__dirname + '/entity/*.js'],
  synchronize: true,
  logging: false
});

module.exports = { AppDataSource };
import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
// import { TypeOrmModule } from '@nestjs/typeorm';
// import { ConfigModule, ConfigService } from '@nestjs/config'; // For configuration

@Module({
  imports: [
    // ConfigModule.forRoot({ isGlobal: true }), // Example: load .env files
    // TypeOrmModule.forRootAsync({
    //   imports: [ConfigModule],
    //   useFactory: (configService: ConfigService) => ({
    //     type: 'postgres',
    //     url: configService.get('DATABASE_URL'), // From .env
    //     entities: [__dirname + '/../**/*.entity{.ts,.js}'],
    //     synchronize: true, // true for dev, false for prod
    //     // autoLoadEntities: true, // Alternative to entities path
    //   }),
    //   inject: [ConfigService],
    // }),
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {} 
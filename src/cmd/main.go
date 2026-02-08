package main

import (
	"context"
	"fmt"

	"saas/src/bootstrap"
	general_middleware "saas/src/general/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	router := gin.Default()
	config := bootstrap.InitConfig()
	generalMiddleware := general_middleware.NewMiddleware(config)
	
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", 
		config.Database.User, config.Database.Password, 
		config.Database.Host, config.Database.Port,
		config.Database.Name)
	
	pool,err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	
	dbManager := bootstrap.NewDBManager(pool)
	
	
	router.Use(generalMiddleware.SubDomainMiddleware()).GET("/", func(c *gin.Context) {
		tenant := c.MustGet("tenant").(string)
		conn, err := dbManager.GetConnForSchema(context.Background(), tenant)
		if err != nil {
			panic(err)
		}
		defer conn.Release()
		// create user table
		_, err = conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, email VARCHAR(255) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL)")
		if err != nil {
			panic(err)
		}
		
		r := fmt.Sprintf("Hello %s",tenant)
		c.String(200, r)
	})
	router.Run(fmt.Sprintf("%s:%s", config.ServerConfig.Host, config.ServerConfig.Port))
}
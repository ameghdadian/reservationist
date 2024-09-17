package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/ameghdadian/service/business/data/dbmigrate"
	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ardanlabs/conf/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/open-policy-agent/opa/rego"
)

var (
	//go:embed rego/authentication.rego
	opaAuthentication string
)

var command string

func init() {
	flag.StringVar(&command, "command", "", "Valid commands: (migrateseed, gentoken, genkey)")
}

func main() {
	flag.Parse()
	// _, err := genkey()
	var err error
	switch command {
	case "migrateseed":
		err = migrateSeed()
	case "gentoken":
		err = gentoken()
	case "genkey":
		_, err = genkey()
	default:
		log.Fatalln(errors.New("unrecognized command"))
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func migrateSeed() error {
	var cfg struct {
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:database-service.reservations-system.svc.cluster.local"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
	}

	const prefix = "RESERVATIONS"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		return fmt.Errorf("parsing config: %w", err)
	}

	dbConfig := db.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	}

	db, err := db.Open(dbConfig)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbmigrate.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")

	// --------------------------------------------------------------------

	if err := dbmigrate.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	fmt.Println("seed data complete")
	return nil
}

func gentoken() error {

	file, err := os.Open("zarf/keys/963df661-d92e-4991-b519-77d838a21705.pem")
	if err != nil {
		return fmt.Errorf("opening key file: %w", err)
	}
	defer file.Close()

	// Limit pem file size to 1 megabyte. This should be reasonable for
	// almost any PEM file and prevents shenanigans like linking the file
	// to /dev/random or something llike that.
	pemData, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("reading auth private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return fmt.Errorf("parsing auth private key: %w", err)
	}

	// To generate a token, we need a set of claims. In this case, we only define
	// subject, user and the roles the user has and token expiration time.
	//
	// Reminder:
	// iss(issuer): Issuer of the JWT
	// sub(subject): Subject of the JWT(the user)
	// aud(audience): Recipient for which the JWT is intended
	// exp(expiration time): Time after which the JWT expires
	// nbf(not before time): Time before which the JWT must not be accepted for further processing
	// iat(issued at time): Time at which the JWT was issued; can be used to determine age of JWT
	// jti(JWT ID): Unique identifer; can be used to prevent the JWT from being replayed(allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string `json:"roles"`
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1234567890",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)

	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "963df661-d92e-4991-b519-77d838a21705"

	str, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println("********************************")
	fmt.Println(str)
	fmt.Println("********************************")

	// --------------------------------------------------------------------

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}

	var claims2 struct {
		jwt.RegisteredClaims
		Roles []string `json:"roles"`
	}

	tkn, err := parser.ParseWithClaims(str, &claims2, keyFunc)
	if err != nil {
		return fmt.Errorf("parsing token: %w", err)
	}

	if !tkn.Valid {
		return errors.New("signature failed")
	}

	fmt.Println("SIGNATURE VALIDATED")
	fmt.Printf("%#v\n", claims2)
	fmt.Println("*********************")

	// --------------------------------------------------------------------

	var claims3 struct {
		jwt.RegisteredClaims
		Roles []string `json:"roles"`
	}

	_, _, err = parser.ParseUnverified(str, &claims3)
	if err != nil {
		return fmt.Errorf("error parsing token unver: %w", err)
	}

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshalling public key: %w", err)
	}

	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	var b bytes.Buffer

	if err := pem.Encode(&b, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	input := map[string]any{
		"Key":   b.String(),
		"Token": str,
	}

	if err := opaPolicyEvaluation(context.Background(), opaAuthentication, input); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("SIGNATURE VALIDATED BY REGO")
	fmt.Println("**************************")

	return nil
}

func genkey() (*rsa.PrivateKey, error) {
	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}

	// Create a file for the private key information in PEM format.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return nil, fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return nil, fmt.Errorf("encoding to private file: %w", err)
	}

	publicFile, err := os.Create("public.pem")
	if err != nil {
		return nil, fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshalling public key: %w", err)
	}

	publicPem := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	if err := pem.Encode(publicFile, &publicPem); err != nil {
		return nil, fmt.Errorf("encoding to public file: %w", err)

	}

	fmt.Println("private and public key files generated")

	return privateKey, nil
}

func opaPolicyEvaluation(ctx context.Context, opaPolicy string, input any) error {
	const opaPackage = "me.rego"
	const rule = "auth"

	// Refer to this for how to access variables using global data variable.
	// https://www.openpolicyagent.org/docs/latest/#complete-rules
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaPolicy),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%#v] ok[%v]", results, ok)
	}

	return nil
}

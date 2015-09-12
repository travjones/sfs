package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/jmoiron/sqlx"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/unrolled/render"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// context
type ctx struct {
	db *sqlx.DB
	r  *render.Render
}

// model
type Country struct {
	CountryID int    `db:"country_id"`
	Name      string `db:"name"`
}

type Supporter struct {
	SupporterID int    `db:"supporter_id"`
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	CountryID   string `db:"country_id"`
	CountryName string `db:"name"`
}

// view
type View struct {
	Title string
	Name  string
	Error string
	Data  interface{}
}

func (c *ctx) NewSupporter(w http.ResponseWriter, r *http.Request) {
	var id string // use a map instead of two slices boiii
	var ids []string
	var country string
	var countries []string

	data := struct {
		IDs       []string
		Countries []string
	}{
		ids,
		countries,
	}

	v := View{
		"New Supporter",
		"newsupporter",
		"",
		data,
	}

	// query db for country_id and name
	query := "select country_id, name from country"
	rows, err := c.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &country)
		if err != nil {
			log.Fatal(err)
		}
		ids = append(ids, id)
		countries = append(countries, country)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	data.IDs = ids
	data.Countries = countries
	v.Data = data

	c.r.HTML(w, http.StatusOK, "newsupporter", v)
}

func (c *ctx) NewSupporterPost(w http.ResponseWriter, r *http.Request) {
	var id string // use a map instead of two slices boiii
	var ids []string
	var country string
	var countries []string

	data := struct {
		IDs       []string
		Countries []string
		First     string
		Last      string
		Country   string
	}{
		ids,
		countries,
		r.FormValue("firstName"),
		r.FormValue("lastName"),
		r.FormValue("country"),
	}

	v := View{
		"New Supporter",
		"newsupporter",
		"",
		data,
	}

	// query db for country_id and name
	query := "select country_id, name from country"
	rows, err := c.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &country)
		if err != nil {
			log.Fatal(err)
		}
		ids = append(ids, id)
		countries = append(countries, country)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	data.IDs = ids
	data.Countries = countries
	v.Data = data

	// post data from form into db
	q := "insert into supporter (first_name, last_name, country_id) values (?, ?, ?);"
	_, err = c.db.Exec(q, data.First, data.Last, data.Country)
	if err != nil {
		v.Error = "Could not add supporter."
		log.Fatal(err)
		c.r.HTML(w, http.StatusOK, "newsupporter", v)
	}

	// redirect to /love
	http.Redirect(w, r, "/love", http.StatusFound)
}

func (c *ctx) ShowSupporters(w http.ResponseWriter, r *http.Request) {

	// most recent supporters select fn, ln, country from supporter oder by supporter_id desc limit 7
	supporters := []Supporter{}

	mostRecentQ := `select s.first_name, s.last_name, s.country_id, c.name
					from supporter as s
					inner join country as c
					on s.country_id=c.country_id
					order by supporter_id desc limit 7;`
	c.db.Select(&supporters, mostRecentQ)

	// most supportive countries

	data := struct {
		Supporters []Supporter
	}{
		supporters,
	}

	v := View{
		"New Supporter",
		"newsupporter",
		"",
		data,
	}

	v.Data = data

	c.r.HTML(w, http.StatusOK, "showsupporters", v)
}

func main() {
	// render
	ren := render.New(render.Options{
		Layout:        "shared/layout",
		IndentJSON:    true,
		IsDevelopment: false,
	})

	//init db
	db, err := sqlx.Connect("mysql", "root:blahblah92@tcp(127.0.0.1:3306)/sfs")
	if err != nil {
		log.Print(err)
		log.Print("Error initializing database...")
	}
	defer db.Close()

	// setup context
	c := ctx{db, ren}

	// setup negroni
	n := negroni.New()

	// setup router
	router := mux.NewRouter()

	// routes
	router.HandleFunc("/", c.NewSupporter).Methods("GET")
	router.HandleFunc("/", c.NewSupporterPost).Methods("POST")
	router.HandleFunc("/love", c.ShowSupporters).Methods("GET")

	//n.Use(NewRecovery(false))
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	n.Use(negroni.NewStatic(http.Dir("public")))

	n.UseHandler(router)

	n.Run(fmt.Sprint(":", os.Getenv("PORT")))
}

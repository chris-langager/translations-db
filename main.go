package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/chris-langager/translationsdb/translations"
	_ "github.com/mattn/go-sqlite3"
)

// var events = []Event{}

// func NewApp(events ...Event) *App {
// 	app := &App{
// 		ProjectsById: map[string]*Project{},
// 		History:      []string{},
// 	}

// 	app.Reduce(events...)

// 	return app
// }

// func NewProject(events ...Event) *Project {
// 	project := &Project{
// 		History: []string{},
// 	}

// 	project.Reduce(events...)

// 	return project
// }

// func (app *App) Reduce(events ...Event) {
// 	for _, event := range events {
// 		switch e := event.(type) {
// 		case ProjectCreated:
// 			project := &Project{}
// 			project.Reduce(event)
// 			app.ProjectsById[e.Payload.Id] = project
// 		case KeyCreated, KeyDeleted:
// 			app.ProjectsById[event.GetAggregateId()].Reduce(event)
// 		}

// 		h, _ := json.MarshalIndent(event, "", "  ")
// 		app.History = append([]string{string(h)}, app.History...)
// 	}
// }

// func (o *Project) Reduce(events ...Event) {
// 	for _, event := range events {
// 		switch e := event.(type) {
// 		case ProjectCreated:
// 			*o = Project{
// 				Id:       e.Payload.Id,
// 				Name:     e.Payload.Name,
// 				KeysById: map[string]*Key{},
// 			}
// 		case KeyCreated:
// 			o.KeysById[e.Payload.Id] = &Key{
// 				Id: e.Payload.Id,
// 			}
// 		}

// 		h, _ := json.MarshalIndent(event, "", "  ")
// 		o.History = append([]string{string(h)}, o.History...)
// 	}
// }

func main() {
	db, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer db.Close()

	eventStore := translations.NewInMemoryEventStore()
	// projectList := translations.NewInMemoryProjectList()

	createProject := translations.NewCommandPipeline(db, translations.CreateProject(), eventStore)
	createKey := translations.NewCommandPipeline(db, translations.CreateKey(), eventStore)
	updateTranslation := translations.NewCommandPipeline(db, translations.UpdateTranslation(), eventStore)

	router := http.NewServeMux()

	router.HandleFunc("POST /translations", func(w http.ResponseWriter, r *http.Request) {
		projectId := r.FormValue("project-id")
		keyId := r.FormValue("key-id")
		id := r.FormValue("id")
		value := r.FormValue("value")
		err := updateTranslation(r.Context(), translations.UpdateTranslationInput{
			ProjectId: projectId,
			KeyId:     keyId,
			Id:        id,
			Value:     value,
		})
		if err != nil {
			panic(err)
		}

		project, err := translations.GetProject(r.Context(), eventStore, projectId)
		if err == translations.ErrorNotFound {
			RenderHtml(w, "fourOhFour.html", nil)
			return
		}
		if err != nil {
			panic(err)
		}

		RenderHtml(w, "translationForm.html", project.KeysById[keyId].TranslationsById[id])
	})

	router.HandleFunc("POST /keys", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		projectId := r.FormValue("project-id")
		err := createKey(r.Context(), translations.CreateKeyInput{
			Id:        id,
			ProjectId: projectId,
		})
		if err != nil {
			panic(err)
		}

		project, err := translations.GetProject(r.Context(), eventStore, projectId)
		if err == translations.ErrorNotFound {
			RenderHtml(w, "fourOhFour.html", nil)
			return
		}
		if err != nil {
			panic(err)
		}

		RenderHtml(w, "project.html", project)
	})

	router.HandleFunc("POST /projects", func(w http.ResponseWriter, r *http.Request) {
		err := createProject(r.Context(), translations.CreateProjectInput{
			Name: r.FormValue("name"),
		})
		if err != nil {
			//TODO: validation error
			panic(err)
		}

		projectList, err := translations.GetProjectList(r.Context(), eventStore)
		if err != nil {
			panic(err)
		}
		RenderHtml(w, "newProjectForm.html", nil)
		RenderHtml(w, "projects.html", projectList.ProjectsById)
		RenderHtml(w, "history.html", projectList.History)
	})

	router.HandleFunc("GET /project/{id}", func(w http.ResponseWriter, r *http.Request) {
		project, err := translations.GetProject(r.Context(), eventStore, r.PathValue("id"))
		if err == translations.ErrorNotFound {
			RenderHtml(w, "fourOhFour.html", nil)
			return
		}
		if err != nil {
			panic(err)
		}

		RenderHtml(w, "projectPage.html", project)
	})

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		var projectList translations.ProjectList
		translations.ReduceWith(r.Context(), &projectList, eventStore.NewGenerator())

		RenderHtml(w, "index.html", projectList)
	})

	fmt.Println("listinging on port 3000...")
	panic(http.ListenAndServe(":3000", router))
}

// TODO: split behavior on local or server
func RenderHtml(wr io.Writer, name string, data any) {
	t, err := template.ParseGlob("**/*.html")
	if err != nil {
		panic(err)
	}

	// load the page we want last (last write wins with the "content" block)
	t, err = t.ParseGlob(fmt.Sprintf("**/%s", name))
	if err != nil {
		panic(err)
	}

	err = t.ExecuteTemplate(wr, name, data)
	if err != nil {
		panic(err)
	}
}

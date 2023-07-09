package main

import (
	"fmt"
	"time"

	"github.com/romshark/taskhub/graph"
	"github.com/romshark/taskhub/graph/model"
	"github.com/romshark/taskhub/slices"
)

func makeData(r *graph.Resolver) {
	// Create users
	userSWEBE_RyanLindsey := &model.User{
		DisplayName: "Ryan Lindsey",
		Role:        "SWE Backend",
		Location:    "NYC ❤️",
	}
	userSWEBE_SamuelBurton := &model.User{
		DisplayName: "Samuel Burton",
		Role:        "SWE Backend",
		Location:    "Boston, US",
	}
	userPLE_ValentinoBonaventura := &model.User{
		DisplayName: "Valentino Bonaventura",
		Role:        "Platform Engineer",
		Location:    "Rome, Italy",
	}
	userOPS_NikitaMykalay := &model.User{
		DisplayName: "Nikita Mykalay",
		Role:        "Ops",
		Location:    "Boston",
	}
	userSWEFR_FrankBohn := &model.User{
		DisplayName: "Frank Bohn",
		Role:        "SWE Frontend",
		Location:    "München",
	}
	userSWEFR_AlmaWarner := &model.User{
		DisplayName: "Alma Warner",
		Role:        "SWE Frontend",
		Location:    "Washington D.C.",
	}
	userSWEBE_DanielLowitz := &model.User{
		DisplayName: "Daniel Lowitz",
		Role:        "SWE Backend",
		Location:    "Washington D.C.",
	}
	userSRE_IvoGarrette := &model.User{
		DisplayName: "Ivo Garrette",
		Role:        "Site-Reliability Engineer",
		Location:    "Munich, Germany",
	}
	usersDTA_Oleksiy := &model.User{
		DisplayName: "Oleksiy Gavrilyuk",
		Role:        "Data Analyst",
		Location:    "Munich",
	}
	usersUXD_EstefaniaPerez := &model.User{
		DisplayName: "Estefania Pérez",
		Location:    "Madrid, Spain",
	}

	// Create line managers
	userPM_AnneWilliams := &model.User{
		DisplayName: "Anne Williams",
		Role:        "Project Manager",
		Location:    "New York",
	}
	userPM_JamesHunter := &model.User{
		DisplayName: "James Hunter",
		Role:        "Project Manager",
		Location:    "Munich",
	}

	// Create BI users
	userBI_RonanMirella := &model.User{
		DisplayName: "Ronan Mirella",
		Role:        "Business Analyst",
		Location:    "Washington D.C.",
	}

	// Create Marketing
	userMAR_AshleyRice := &model.User{
		DisplayName: "Ashley Rice",
		Role:        "Marketing Director",
		Location:    "Washington D.C.",
	}

	// Create HR users
	userHR_AdrienneClement := &model.User{
		DisplayName: "Adrienne Clement",
		Role:        "HR",
		Location:    "UK - London",
	}

	// Create upper management users
	userCTO_MarcCarlson := &model.User{
		DisplayName: "Marc Carlson",
		Role:        "CTO",
		Location:    "New York City",
	}
	userCFO_TabbyWalters := &model.User{
		DisplayName: "Tabby Walters",
		Role:        "CFO",
		Location:    "New York",
	}
	userCOO_BrittanyEmile := &model.User{
		DisplayName: "Brittany Emile",
		Role:        "COO",
		Location:    "New York",
	}
	userCEO_CedricMaude := &model.User{
		DisplayName: "Cedric Maude",
		Role:        "CEO",
		Location:    "New York City",
	}

	r.Users = []*model.User{
		userSWEBE_RyanLindsey,
		userSWEBE_SamuelBurton,
		userPLE_ValentinoBonaventura,
		userOPS_NikitaMykalay,
		userSWEFR_FrankBohn,
		userSWEFR_AlmaWarner,
		userSWEBE_DanielLowitz,
		userSRE_IvoGarrette,
		usersDTA_Oleksiy,
		usersUXD_EstefaniaPerez,
		userPM_AnneWilliams,
		userPM_JamesHunter,
		userBI_RonanMirella,
		userMAR_AshleyRice,
		userHR_AdrienneClement,
		userCTO_MarcCarlson,
		userCFO_TabbyWalters,
		userCOO_BrittanyEmile,
		userCEO_CedricMaude,
	}

	// Set user IDs
	for _, u := range r.Users {
		u.ID = "user_" + graph.MakeID(u.DisplayName)
	}

	// Set subordinate->manager relations
	userSWEBE_RyanLindsey.Manager = userPM_AnneWilliams
	userSWEBE_SamuelBurton.Manager = userPM_AnneWilliams
	userPLE_ValentinoBonaventura.Manager = userPM_AnneWilliams
	userOPS_NikitaMykalay.Manager = userPM_AnneWilliams
	userSWEFR_FrankBohn.Manager = userPM_AnneWilliams

	userSWEFR_AlmaWarner.Manager = userPM_JamesHunter
	userSWEBE_DanielLowitz.Manager = userPM_JamesHunter
	userSRE_IvoGarrette.Manager = userPM_JamesHunter
	usersDTA_Oleksiy.Manager = userPM_JamesHunter
	usersUXD_EstefaniaPerez.Manager = userPM_JamesHunter

	userPM_AnneWilliams.Manager = userCTO_MarcCarlson
	userPM_JamesHunter.Manager = userCTO_MarcCarlson

	userBI_RonanMirella.Manager = userCOO_BrittanyEmile
	userMAR_AshleyRice.Manager = userCOO_BrittanyEmile
	userHR_AdrienneClement.Manager = userCOO_BrittanyEmile

	userCTO_MarcCarlson.Manager = userCEO_CedricMaude
	userCFO_TabbyWalters.Manager = userCEO_CedricMaude
	userCOO_BrittanyEmile.Manager = userCEO_CedricMaude

	// Set manager->subordinates relations
	for _, u1 := range r.Users {
		for _, u2 := range r.Users {
			if u1.Manager == u2 {
				u2.Subordinates = slices.AppendUnique(u2.Subordinates, u1)
			}
		}
	}

	// Create projects
	projectCoreMigration := &model.Project{
		Name:   "Core Migration",
		Slug:   "CORM",
		Owners: []*model.User{userPM_AnneWilliams},
	}
	projectVendorPlatform := &model.Project{
		Name:   "Vendor Platform",
		Slug:   "VENP",
		Owners: []*model.User{userPM_JamesHunter},
	}
	projectPlatformUpgrade := &model.Project{
		Name:   "Platform Upgrade",
		Slug:   "PLUG",
		Owners: []*model.User{userPM_JamesHunter},
	}

	r.Projects = []*model.Project{
		projectCoreMigration,
		projectVendorPlatform,
		projectPlatformUpgrade,
	}

	// Set project IDs
	for _, p := range r.Projects {
		p.ID = "project_" + graph.MakeID(p.Name)
	}

	const (
		taskTagBackend  = "backend"
		taskTagFrontend = "frontend"
		taskTagDatabase = "database"
		taskTagOps      = "ops"
	)

	// Create tasks for Project: Core Migration
	task1 := &model.Task{
		Title: "Implement database migration",
		Description: ptr(`Define and implement a comprehensive database migration` +
			`strategy for the core system. Assess the existing database structure, ` +
			`schema, and data dependencies. Plan and execute the migration process, ` +
			`ensuring data integrity and minimal downtime. Collaborate with the` +
			`backend team to handle data transformations and perform necessary ` +
			`optimizations during the migration.`),
		Status:   model.TaskStatusInProgress,
		Creation: time.Now().AddDate(0, 0, -4),
		Due:      ptr(time.Now().AddDate(0, 0, -1)),
		Tags:     []string{taskTagBackend, taskTagDatabase},
		Project:  projectCoreMigration,
		Assignees: []*model.User{
			userOPS_NikitaMykalay,
			userPLE_ValentinoBonaventura,
		},
		Reporters: []*model.User{userPM_AnneWilliams},
	}
	task2 := &model.Task{
		Title: "Design dashboard components",
		Description: ptr(`Create detailed design mockups and interactive wireframes ` +
			`for the frontend components of the application. Incorporate the latest ` +
			`UX/UI principles and best practices to provide an intuitive and visually` +
			` appealing user experience. Consider accessibility requirements and` +
			`ensure cross-browser compatibility. Collaborate closely with the frontend ` +
			`developers to ensure seamless implementation of the designs.`),
		Status:    model.TaskStatusTodo,
		Creation:  time.Now(),
		Due:       nil,
		Tags:      []string{taskTagFrontend},
		Project:   projectCoreMigration,
		Assignees: []*model.User{userSWEFR_FrankBohn, usersUXD_EstefaniaPerez},
		Reporters: []*model.User{userPM_AnneWilliams},
	}

	// Create tasks for Project: Vendor Platform
	task3 := &model.Task{
		Title:       "Fix timeouts, optimize handler performance",
		Description: ptr(`Some handlers are causing timeouts.`),
		Status:      model.TaskStatusInProgress,
		Creation:    time.Now().AddDate(0, 0, -1),
		Due:         ptr(time.Now().AddDate(0, 0, 3)),
		Tags:        []string{taskTagBackend},
		Project:     projectVendorPlatform,
		Assignees:   []*model.User{userSWEBE_DanielLowitz, userSRE_IvoGarrette},
		Reporters:   []*model.User{userPLE_ValentinoBonaventura},
	}
	task4 := &model.Task{
		Title: "Refactor frontend code",
		Description: ptr(`Refactor the frontend codebase to enhance scalability, ` +
			`maintainability, and code reusability. Identify areas of code` +
			` duplication complex logic, and poor architectural patterns. ` +
			`Restructure the codebase using modular design principles, ` +
			`separating concerns into reusable components. ` +
			`Apply best practices, such as code commenting, code ` +
			`organization, and consistent naming conventions, ` +
			`to improve code readability and maintainability.`),
		Status:    model.TaskStatusDone,
		Creation:  time.Now().Add(-(time.Hour * 2)),
		Due:       ptr(time.Now().AddDate(0, 0, 2)),
		Tags:      []string{"frontend"},
		Project:   projectVendorPlatform,
		Assignees: []*model.User{usersUXD_EstefaniaPerez},
		Reporters: []*model.User{userPM_JamesHunter},
	}

	// Create tasks for Project: Platform Upgrade
	task5 := &model.Task{
		Title:       "Migrate to latest platform version",
		Description: nil,
		Status:      model.TaskStatusTodo,
		Creation:    time.Now().Add(-(time.Hour * 1)),
		Due:         ptr(time.Now().AddDate(0, 0, 4)),
		Tags:        []string{taskTagBackend},
		Project:     projectPlatformUpgrade,
		Assignees:   []*model.User{userSRE_IvoGarrette},
		Reporters:   []*model.User{userPM_JamesHunter},
	}
	task6 := &model.Task{
		Title:       "Implement deployment automation",
		Description: nil,
		Status:      model.TaskStatusInProgress,
		Creation:    time.Now().Add(-(time.Minute * 5)),
		Due:         ptr(time.Now().AddDate(0, 0, 6)),
		Tags:        []string{taskTagOps},
		Project:     projectPlatformUpgrade,
		Assignees:   []*model.User{userOPS_NikitaMykalay},
		Reporters:   []*model.User{userPM_JamesHunter},
	}
	task7 := &model.Task{
		Title:       "Determine upgrade version",
		Description: nil,
		Status:      model.TaskStatusDone,
		Creation:    time.Now().Add(-(time.Minute * 5)),
		Due:         nil,
		Tags:        []string{taskTagOps},
		Project:     projectPlatformUpgrade,
		Assignees:   []*model.User{userOPS_NikitaMykalay},
		Reporters:   []*model.User{userOPS_NikitaMykalay},
	}

	r.Tasks = []*model.Task{
		task1,
		task2,
		task3,
		task4,
		task5,
		task6,
		task7,
	}

	for i, t := range r.Tasks {
		// Set task IDs
		t.ID = fmt.Sprintf("task_%s_%d", graph.MakeID(t.Project.Slug), i)
	}
}

func ptr[T any](v T) *T { return &v }

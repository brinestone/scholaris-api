package institutions

import "encore.dev/cron"

var _ = cron.NewJob("academic-year-creation", cron.JobConfig{
	Title:    "Create Academic Years",
	Schedule: "0 0 * * *", // ! Every midnight
	Endpoint: AutoCreateAcademicYears,
})

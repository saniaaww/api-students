package model

type Supervisor struct {
    ID              int    `json:"id"`
    Name            string `json:"name"`
    Quota           int    `json:"quota"`
    SupervisedCount int    `json:"supervised_count"`
    Version         int    `json:"version"`
}

type AssignSupervisorRequest struct {
    StudentID int `json:"student_id"`
}

type SwitchSupervisorRequest struct {
    StudentID       int `json:"student_id"`
    NewSupervisorID int `json:"new_supervisor_id"`
}

type UpdateSupervisorQuotaRequest struct {
    Quota   int `json:"quota"`
    Version int `json:"version"`
}

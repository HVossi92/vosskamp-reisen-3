package services

import (
	"database/sql"
	"fmt"
	"vosskamp-reisen-3/models"
)

func FetchTaskById(db *sql.DB, id int) (*models.Task, error) {
	query := "SELECT * FROM tasks WHERE id = ? LIMIT 1"
	row := db.QueryRow(query, id)

	var task models.Task
	err := row.Scan(&task.Id, &task.Name, &task.Done, &task.Created, &task.Updated)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func FetchTasks(db *sql.DB) ([]models.Task, error) {
	query := "SELECT * FROM tasks"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.Id, &task.Name, &task.Done, &task.Created, &task.Updated)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func InsertTask(db *sql.DB, taskName string) error {
	query := "INSERT INTO tasks (name) VALUES (?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskName)
	if err != nil {
		return err
	}

	return nil
}

func UpdateTask(db *sql.DB, task models.Task) (*models.Task, error) {
	query := "UPDATE tasks SET name = ?, done = ? WHERE id = ? RETURNING *"
	var updatedTask models.Task
	err := db.QueryRow(query, task.Name, task.Done, task.Id).Scan(&updatedTask.Id, &updatedTask.Name, &updatedTask.Done, &updatedTask.Created, &updatedTask.Updated)

	if err != nil {
		if err == sql.ErrNoRows {
			// No rows were updated
			return nil, fmt.Errorf("no task found with id %d", task.Id)
		}
		return nil, err
	}

	return &updatedTask, nil
}

func RemoveTask(db *sql.DB, taskId int) error {
	query := "DELETE FROM tasks WHERE id = (?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskId)
	if err != nil {
		return err
	}

	return nil
}

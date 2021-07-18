package main

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Poll struct {
	Name     string   `json:"name"`
	Variants []string `json:"variants"`
}

type VotingSheet struct {
	Poll_id   int `json:"poll_id"`
	Choice_id int `json:"choice_id"`
}

type Field struct {
	Choice_id   int    `json:"choice_id" db:"id_choice"`
	Choice_name string `json:"choice_name" db:"name_choice"`
	Voters      int    `json:"voters" db:"voters"`
}

type Report struct {
	result []Field
}

type Storage interface {
	Add(p *Poll) (int, error)
	Vote(poll_id int, choice_id int) (int, error)
	ReportBack(poll_id int) (Report, error)
}

type MemoryStorage struct {
	db *sqlx.DB
	sync.Mutex
}

func NewMemoryStorage() *MemoryStorage {
	db := sqlx.MustConnect("postgres", "user=postgres password=admin dbname=avito_intern sslmode=disable")
	return &MemoryStorage{
		db: db,
	}
}

func (s *MemoryStorage) Add(p *Poll) (int, error) {
	if err := checkName(p.Name); err != nil {
		return -1, err
	}
	if err := checkVariants(p.Variants); err != nil {
		return -1, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}

	var newPollId int
	tx.QueryRow("INSERT INTO polls (name_poll) VALUES ($1) RETURNING ID_POLL", p.Name).Scan(&newPollId)
	if newPollId == 0 {
		tx.Rollback()
		return -1, fmt.Errorf("Something wrong with insert into polls.")
	}

	for i, value := range p.Variants {
		_, err := tx.Exec("INSERT INTO counter (id_poll, id_choice, name_choice, voters) VALUES ($1, $2, $3, $4)", newPollId, i+1, value, 0)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
	}
	tx.Commit()
	return newPollId, nil
}

func (s *MemoryStorage) Vote(poll_id int, choice_id int) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}

	var voters int
	tx.QueryRow("UPDATE counter SET voters = voters + 1 WHERE id_poll = $1 AND id_choice = $2 RETURNING voters", poll_id, choice_id).Scan(&voters)
	if voters == 0 {
		tx.Rollback()
		return -1, fmt.Errorf("Something wrong with the voting. Your vote is not counted.")
	}

	tx.Commit()
	return voters, nil
}

func (s *MemoryStorage) ReportBack(poll_id int) (Report, error) {

	r := Report{result: make([]Field, 0)}

	err := s.db.Select(&r.result, "SELECT id_choice, name_choice, voters FROM counter WHERE id_poll = $1", poll_id)

	if err != nil {
		return r, err
	}

	return r, nil
}

/*Корректность имени.*/

func checkName(str string) error {
	if len(str) == 0 {
		return fmt.Errorf("Name is too short.")
	}
	if len(str) > 100 {
		return fmt.Errorf("Name is too long. Change it.")
	}
	return nil
}

/*Корректность вариантов ответа.
Возможно, имело смысл приводить варианты ответа в один формат:
нижний регистр, один пробел между словами и так далее;
но пока без этого.*/

func checkVariants(variants []string) error {
	if len(variants) <= 1 {
		return fmt.Errorf("Count of variants is too small.")
	}
	if len(variants) > 16000 {
		return fmt.Errorf("Count of variants is too big.")
	}
	unique := make(map[string]int, len(variants))
	for i, value := range variants {
		if len(value) == 0 {
			return fmt.Errorf("Check the variant %d'th.", i)
		}
		if len(value) > 100 {
			return fmt.Errorf("Check the variant %d'th.", i)
		}
		if unique[value] == 1 {
			return fmt.Errorf("Сheck the uniqueness of the answer options.")
		}
		unique[value] = 1
	}

	return nil
}

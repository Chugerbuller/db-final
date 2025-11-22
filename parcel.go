package main	

import (
	"database/sql"
	"log"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {

	query := `INSERT INTO parcel (client, status, address, created_at) 
              VALUES (:client, :status, :address, :created_at)`
	res, err := s.db.Exec(query,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt),
	)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), nil

}

func (s ParcelStore) Get(number int) (Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE number = :number;"
	row := s.db.QueryRow(query, sql.Named("number", number))
	p := Parcel{}

	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE client = :client;"
	rows, err := s.db.Query(query, sql.Named("client", client))
	if err != nil {
		return []Parcel{}, err
	}
	defer rows.Close()

	var res []Parcel

	for rows.Next() {
		p := Parcel{}
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return res, err
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	query := "UPDATE parcel set status = :status where number = :number;"
	_, err := s.db.Exec(query,
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		log.Printf("ERROR: Failed to set parcel %v: %v", number, err)
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	query := "UPDATE parcel SET address = :address WHERE number = :number AND status = :status;"
	_, err := s.db.Exec(query,
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}
	return nil

}

func (s ParcelStore) Delete(number int) error {
	query := "DELETE FROM parcel WHERE number = :number AND status = :status;"
	_, err := s.db.Exec(query,
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}
	return nil
}

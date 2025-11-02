package main

import (
	"database/sql"
	"errors"
	"log"
)

var (
	errWrongStatus = errors.New("wrong status")
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
		log.Printf("ERROR: Failed to insert parcel %v: %v", p, err)
		return -1, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		log.Printf("ERROR: Failed to get last insert ID for client %d: %v", p.Client, err)
		return -1, err
	}

	return int(lastID), nil

}

func (s ParcelStore) Get(number int) (Parcel, error) {
	query := "SELECT * FROM parcel WHERE number = :number;"
	row := s.db.QueryRow(query, sql.Named("number", number))
	p := Parcel{}

	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		log.Printf("ERROR: Failed to get parcel %v: %v", number, err)
		return Parcel{}, err
	}

	if err = row.Err(); err != nil {
		log.Printf("ERROR: Failed to read columns from parcel %v: %v", number, err)
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := "SELECT * FROM parcel WHERE client = :client;"
	rows, err := s.db.Query(query, sql.Named("client", client))
	if err != nil {
		log.Printf("ERROR: Failed to get parcels by client %v: %v", client, err)
		return []Parcel{}, err
	}
	defer rows.Close()

	var res []Parcel

	for rows.Next() {
		p := Parcel{}
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			log.Printf("ERROR: Failed to scan parcels by client %v: %v", client, err)
			return res, err
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Failed to scan parcels by client %v: %v", client, err)
		return res, err
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
	parcel, err := s.Get(number)
	if err != nil {
		log.Printf("ERROR: Failed to get parcel %v: %v", number, err)
		return err
	}
	if parcel.Status != ParcelStatusRegistered {
		log.Printf("ERROR: Parcel %v has wrong status %v", parcel.Number, parcel.Status)
		return errWrongStatus
	}
	query := "UPDATE parcel SET address = :address WHERE number = :number;"
	_, err = s.db.Exec(query,
		sql.Named("address", address),
		sql.Named("number", number))
	if err != nil {
		log.Printf("ERROR: Failed to set parcel %v: %v", parcel.Number, err)
		return err
	}
	return nil

}

func (s ParcelStore) Delete(number int) error {
	parcel, err := s.Get(number)
	if err != nil {
		log.Printf("ERROR: Failed to get parcel %v: %v", number, err)
		return err
	}
	if parcel.Status != ParcelStatusRegistered {
		log.Printf("ERROR: Parcel %v has wrong status: %v", parcel.Number, parcel.Status)
		return errWrongStatus
	}

	query := "DELETE FROM parcel WHERE number = :number;"
	_, err = s.db.Exec(query, sql.Named("number", number))
	if err != nil {
		log.Printf("ERROR: Failed to delete parcel %v: %s", parcel.Number, err.Error())
		return err
	}
	return nil
}

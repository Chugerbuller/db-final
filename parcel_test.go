package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	parcelId, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = parcelId
	pl, err := store.Get(parcelId)
	require.NoError(t, err)
	assert.Equal(t, parcel, pl)

	err = store.Delete(parcelId)
	require.NoError(t, err)
	_, err = store.Get(parcelId)
	assert.Equal(t, sql.ErrNoRows, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		{
			Client:    1000,
			Status:    ParcelStatusDelivered,
			Address:   "test",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
	for _, parcel := range parcels {
		parcelId, err := store.Add(parcel)
		require.NoError(t, err)
		newAddress := "new test address"
		err = store.SetAddress(parcelId, newAddress)
		switch parcel.Status {
		case ParcelStatusRegistered:
			require.NoError(t, err)
			stored, err := store.Get(parcelId)
			require.NoError(t, err)
			assert.Equal(t, newAddress, stored.Address)
		default:
			assert.Equal(t, err, errWrongStatus)
			stored, err := store.Get(parcelId)
			require.NoError(t, err)
			assert.Equal(t, parcel.Address, stored.Address)
		}

	}

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{getTestParcel(), getTestParcel(), getTestParcel()}
	statuses := []string{ParcelStatusSent, ParcelStatusDelivered, ParcelStatusRegistered}

	for i, status := range statuses {
		parcelId, err := store.Add(parcels[i])
		require.NoError(t, err)

		err = store.SetStatus(parcelId, status)
		switch parcels[i].Status {
		case ParcelStatusRegistered:
			require.NoError(t, err)

		default:
			assert.Equal(t, err, errWrongStatus)
			
		}
		stored, err := store.Get(parcelId)
		require.NoError(t, err)
		assert.Equal(t, status, stored.Status)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Equal(t, len(storedParcels), len(parcels))

	for _, parcel := range storedParcels {
		value, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		assert.Equal(t, parcel, value)
	}
}

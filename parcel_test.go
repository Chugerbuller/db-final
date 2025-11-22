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
	require.ErrorIs(t, sql.ErrNoRows, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	parcelId,err := store.Add(parcel)
	require.NoError(t,err)
	assert.NotZero(t,parcelId)

	newAddress := "new test address"
	err = store.SetAddress(parcelId,newAddress)
	require.NoError(t,err)

	parcel,err = store.Get(parcelId)
	require.NoError(t,err)
	assert.Equal(t,parcel.Address,newAddress)

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	
	parcelId,err := store.Add(parcel)
	require.NoError(t,err)
	assert.NotZero(t,parcelId)

	newStatus := ParcelStatusDelivered
	err = store.SetStatus(parcelId,newStatus)
	require.NoError(t,err)
	parcel,err = store.Get(parcelId)
	require.NoError(t,err)
	assert.Equal(t,parcel.Status,newStatus)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()
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
	assert.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels {
		value, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		assert.Equal(t, parcel, value)
	}
}

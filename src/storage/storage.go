package storage

type Storage interface {
    // Save saves the data to the storage
    Save(data []byte) error
    // Load loads the data from the storage
    Load() ([]byte, error)
}

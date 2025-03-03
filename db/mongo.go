// Package db provides functionality to interact with a MongoDB database for song recognition.
// This package includes methods to store and retrieve song fingerprints, register songs, and manage song collections.
// The MongoDB database is connected to the front-end part of the application, enabling users to interact with the song recognition system.

package db

// MongoClient is a wrapper around the MongoDB client to provide methods for interacting with the song recognition database.
type MongoClient struct {
	client *mongo.Client
}

// NewMongoClient creates a new MongoClient instance and connects to the MongoDB database using the provided URI.
func NewMongoClient(uri string) (*MongoClient, error) {}

// Close disconnects the MongoClient from the MongoDB database.
func (db *MongoClient) Close() error {}

// StoreFingerprints stores song fingerprints in the MongoDB database.
func (db *MongoClient) StoreFingerprints(fingerprints map[uint32]models.Couple) error {}

// GetCouples retrieves song couples from the MongoDB database based on the provided addresses.
func (db *MongoClient) GetCouples(addresses []uint32) (map[uint32][]models.Couple, error) {}

// TotalSongs returns the total number of songs stored in the MongoDB database.
func (db *MongoClient) TotalSongs() (int, error) {}

// RegisterSong registers a new song in the MongoDB database with the provided title, artist, and YouTube ID.
func (db *MongoClient) RegisterSong(songTitle, songArtist, ytID string) (uint32, error) {}

// GetSong retrieves a song from the MongoDB database based on the provided filter key and value.
func (db *MongoClient) GetSong(filterKey string, value interface{}) (s Song, songExists bool, e error) {}

// GetSongByID retrieves a song from the MongoDB database based on the song ID.
func (db *MongoClient) GetSongByID(songID uint32) (Song, bool, error) {}

// GetSongByYTID retrieves a song from the MongoDB database based on the YouTube ID.
func (db *MongoClient) GetSongByYTID(ytID string) (Song, bool, error) {}

		update := bson.M{
			"$push": bson.M{
				"couples": bson.M{
					"anchorTimeMs": couple.AnchorTimeMs,
					"songID":       couple.SongID,
				},
			},
		}
		opts := options.Update().SetUpsert(true)

		_, err := collection.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			return fmt.Errorf("error upserting document: %s", err)
		}
	}

	return nil
}

func (db *MongoClient) GetCouples(addresses []uint32) (map[uint32][]models.Couple, error) {
	collection := db.client.Database("song-recognition").Collection("fingerprints")

	couples := make(map[uint32][]models.Couple)

	for _, address := range addresses {
		// Find the document corresponding to the address
		var result bson.M
		err := collection.FindOne(context.Background(), bson.M{"_id": address}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			}
			return nil, fmt.Errorf("error retrieving document for address %d: %s", address, err)
		}

		// Extract couples from the document and append them to the couples map
		var docCouples []models.Couple
		couplesList, ok := result["couples"].(primitive.A)
		if !ok {
			return nil, fmt.Errorf("couples field in document for address %d is not valid", address)
		}

		for _, item := range couplesList {
			itemMap, ok := item.(primitive.M)
			if !ok {
				return nil, fmt.Errorf("invalid couple format in document for address %d", address)
			}

			couple := models.Couple{
				AnchorTimeMs: uint32(itemMap["anchorTimeMs"].(int64)),
				SongID:       uint32(itemMap["songID"].(int64)),
			}
			docCouples = append(docCouples, couple)
		}
		couples[address] = docCouples
	}

	return couples, nil
}

func (db *MongoClient) TotalSongs() (int, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("songs")
	total, err := existingSongsCollection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

func (db *MongoClient) RegisterSong(songTitle, songArtist, ytID string) (uint32, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("songs")

	// Create a compound unique index on ytID and key, if it doesn't already exist
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"ytID", 1}, {"key", 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := existingSongsCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return 0, fmt.Errorf("failed to create unique index: %v", err)
	}

	// Attempt to insert the song with ytID and key
	songID := utils.GenerateUniqueID()
	key := utils.GenerateSongKey(songTitle, songArtist)
	_, err = existingSongsCollection.InsertOne(context.Background(), bson.M{"_id": songID, "key": key, "ytID": ytID})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, fmt.Errorf("song with ytID or key already exists: %v", err)
		} else {
			return 0, fmt.Errorf("failed to register song: %v", err)
		}
	}

	return songID, nil
}

var mongofilterKeys = "_id | ytID | key"

func (db *MongoClient) GetSong(filterKey string, value interface{}) (s Song, songExists bool, e error) {
	if !strings.Contains(mongofilterKeys, filterKey) {
		return Song{}, false, errors.New("invalid filter key")
	}

	songsCollection := db.client.Database("song-recognition").Collection("songs")
	var song bson.M

	filter := bson.M{filterKey: value}

	err := songsCollection.FindOne(context.Background(), filter).Decode(&song)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Song{}, false, nil
		}
		return Song{}, false, fmt.Errorf("failed to retrieve song: %v", err)
	}

	ytID := song["ytID"].(string)
	title := strings.Split(song["key"].(string), "---")[0]
	artist := strings.Split(song["key"].(string), "---")[1]

	songInstance := Song{title, artist, ytID}

	return songInstance, true, nil
}

func (db *MongoClient) GetSongByID(songID uint32) (Song, bool, error) {
	return db.GetSong("_id", songID)
}

func (db *MongoClient) GetSongByYTID(ytID string) (Song, bool, error) {
	return db.GetSong("ytID", ytID)
}

func (db *MongoClient) GetSongByKey(key string) (Song, bool, error) {
	return db.GetSong("key", key)
}

func (db *MongoClient) DeleteSongByID(songID uint32) error {
	songsCollection := db.client.Database("song-recognition").Collection("songs")

	filter := bson.M{"_id": songID}

	_, err := songsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete song: %v", err)
	}

	return nil
}

func (db *MongoClient) DeleteCollection(collectionName string) error {
	collection := db.client.Database("song-recognition").Collection(collectionName)
	err := collection.Drop(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting collection: %v", err)
	}
	return nil
}

package catTracks

import (
	"encoding/json"
	"fmt"
	"path"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
)

var (
	db         *bolt.DB
	trackKey   = "tracks"
	allBuckets = []string{trackKey, "names", "geohash"}
)

// GetDB is db getter.
func GetDB() *bolt.DB {
	return db
}

func ensureInitBuckets(buckets []string) error {
	err := GetDB().Update(func(tx *bolt.Tx) error {
		var e error
		for _, buck := range buckets {
			fmt.Println("Ensured existance of bucket: ", buck)
			_, e = tx.CreateBucketIfNotExists([]byte(buck))
			if e != nil {
				return e
			}
		}
		return e
	})
	return err
}

// InitBoltDB sets up initial stuff, like the file and necesary buckets
func InitBoltDB() error {
	var err error
	db, err = bolt.Open(path.Join("db", "tracks.db"), 0666, nil)
	if err != nil {
		fmt.Println("Could not initialize Bolt database. ", err)
		return err
	}

	err = ensureInitBuckets(allBuckets)
	if err != nil {
		fmt.Println("Err initing buckets.", err)
	}
	return err
}

//BuildIndexBuckets populates name, lat, and long buckets from main "tracks" (time) bucket.
func BuildIndexBuckets() error {
	var tps []trackPoint.TrackPoint
	var countTps int

	e := db.View(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("names"))
		tx.CreateBucketIfNotExists([]byte("names"))

		tx.DeleteBucket([]byte("geohash"))
		tx.CreateBucketIfNotExists([]byte("geohash"))

		b := tx.Bucket([]byte(trackKey))
		countTps = b.Stats().KeyN

		b.ForEach(func(key, val []byte) error {
			var tp trackPoint.TrackPoint
			json.Unmarshal(val, &tp)
			tps = append(tps, tp)
			return nil
		})
		return nil
	})
	if e != nil {
		fmt.Println("e", e)
		return e
	}

	// update "name"
	fmt.Println("Indexing on names...")
	namebar := pb.StartNew(countTps)

	db.Update(func(txx *bolt.Tx) error {
		bname := txx.Bucket([]byte("names"))
		for _, tp := range tps {
			bByName, _ := bname.CreateBucketIfNotExists([]byte(tp.Name))
			b, e := json.Marshal(tp)
			if e != nil {
				fmt.Println("got err marshaling tp for namer", e)
				return e
			}
			bByName.Put(itob(tp.ID), b)
			namebar.Increment()
		}
		return nil
	})

	namebar.FinishPrint("Finished names.")

	fmt.Println("Indexing on geohash...")
	geobar := pb.StartNew(countTps)
	geobar.ShowFinalTime = true

	// under geohasher keys
	eg := db.Update(func(txx *bolt.Tx) error {
		gb := txx.Bucket([]byte("geohash"))
		for i, tp := range tps {

			hashkey := NewGeoKey(tp)

			if i == 0 {
				fmt.Println("hashkey looks like ", hashkey, string(hashkey), " with len ", len(hashkey))
			}

			b, e := json.Marshal(tp)
			if e != nil {
				fmt.Println("shite rr marshaling tp")
			}

			ep := gb.Put(hashkey, b)
			if ep != nil {
				fmt.Println("NOTSAVE geohash index", ep)
			}
			geobar.Increment()
		}
		return nil
	})
	if eg != nil {
		fmt.Println("eg", eg)
	}
	geobar.FinishPrint("Finished geohashes.")

	// double check.
	var endCountNames, endCountHashes int
	db.View(func(tx *bolt.Tx) error {
		names := tx.Bucket([]byte("names"))
		hashes := tx.Bucket([]byte("geohash"))
		endCountNames, endCountHashes = names.Stats().KeyN, hashes.Stats().KeyN
		return nil
	})

	fmt.Println("Names count: ", endCountNames, "  Hashes count: ", endCountHashes)
	return nil
}

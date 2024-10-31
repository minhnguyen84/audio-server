package outbound

import (
	"audio-server/audiolab"
	"audio-server/utils"
	"fmt"
	externalaudio "github.com/go-audio/audio"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestNoGoroutineLeak vérifie qu'il n'y a pas de fuites de goroutines après l'exécution
func TestNoGoroutineLeak(t *testing.T) {
	// INIT TEST
	initialGoroutines := getGoroutineCount()
	dispatchChan, fileStorage, err := initStorage()
	if err != nil {
		t.Fatalf("Erreur lors de la création de FileStorage: %v", err)
	}

	// GIVEN
	numSessions := 50

	// WHEN
	for i := 0; i < numSessions; i++ {
		sessionId := fmt.Sprintf("Session_%d", i)
		// Envoyer un message d'ouverture
		msgOpen := audiolab.Message{
			SessionId: sessionId,
			Data: audiolab.Data{
				Internal: newMockIntBuffer(),
				External: newMockIntBuffer(),
			},
			IsClosed: false,
		}
		dispatchChan <- msgOpen

		for j := 0; j < 5; j++ {
			msg := audiolab.Message{
				SessionId: sessionId,
				Data: audiolab.Data{
					Internal: newMockIntBuffer(),
					External: newMockIntBuffer(),
				},
				IsClosed: false,
			}
			dispatchChan <- msg
		}

		msgClose := audiolab.Message{
			SessionId: sessionId,
			IsClosed:  true,
		}
		dispatchChan <- msgClose
	}
	close(dispatchChan)
	fileStorage.Close()

	// THEN
	time.Sleep(500 * time.Millisecond)
	finalGoroutines := getGoroutineCount()
	// Autoriser un léger dépassement dû aux goroutines du runtime
	if finalGoroutines > initialGoroutines+1 {
		t.Errorf("Fuite de goroutines détectée. Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	} else {
		t.Logf("Aucune fuite de goroutine détectée. Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	}
}

// TestFileStorage_NoMemoryLeak vérifie l'absence de fuites de goroutines pour FileStorage
func TestFileStorage_NoMemoryLeak(t *testing.T) {
	// INIT
	initialGoroutines := getGoroutineCount()
	dispatchChan, fileStorage, err := initStorage()
	if err != nil {
		t.Fatalf("Erreur lors de la création de FileStorage: %v", err)
	}

	// GIVEN
	numSessions := 100
	var wg sync.WaitGroup
	wg.Add(numSessions)

	// WHEN -> send data to storage
	for i := 0; i < numSessions; i++ {
		go func(i int) {
			defer wg.Done()
			sessionId := fmt.Sprintf("LeakSession_%d", i)
			// Envoyer un message d'ouverture
			msgOpen := audiolab.Message{
				SessionId: sessionId,
				Data: audiolab.Data{
					Internal: newMockIntBuffer(),
					External: newMockIntBuffer(),
				},
				IsClosed: false,
			}
			dispatchChan <- msgOpen

			// Simuler quelques messages intermédiaires
			for j := 0; j < 3; j++ {
				msg := audiolab.Message{
					SessionId: sessionId,
					Data: audiolab.Data{
						Internal: newMockIntBuffer(),
						External: newMockIntBuffer(),
					},
					IsClosed: false,
				}
				dispatchChan <- msg
			}

			// Envoyer un message de fermeture
			msgClose := audiolab.Message{
				SessionId: sessionId,
				IsClosed:  true,
			}
			dispatchChan <- msgClose
		}(i)
	}
	wg.Wait()
	close(dispatchChan)
	fileStorage.Close()

	// THEN
	// Attendre un court instant pour s'assurer que toutes les goroutines ont eu le temps de se terminer
	time.Sleep(500 * time.Millisecond)
	finalGoroutines := getGoroutineCount()
	// Autoriser un léger dépassement dû aux goroutines du runtime
	if finalGoroutines > initialGoroutines+1 {
		t.Errorf("Fuite de goroutines détectée. Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	} else {
		t.Logf("Aucune fuite de goroutine détectée. Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	}
}

// TestFileStorage_SingleSession vérifie la création et la fermeture d'un seul worker
func TestFileStorage_SingleSession(t *testing.T) {
	// INIT
	initialGoroutines := getGoroutineCount()
	dispatchChan, fileStorage, err := initStorage()
	if err != nil {
		t.Fatalf("Erreur lors de la création de FileStorage: %v", err)
	}

	// GIVEN
	sessionId := "SingleSession"

	// WHEN
	msgOpen := audiolab.Message{
		SessionId: sessionId,
		Data: audiolab.Data{
			Internal: newMockIntBuffer(),
			External: newMockIntBuffer(),
		},
		IsClosed: false,
	}
	dispatchChan <- msgOpen

	for i := 0; i < 3; i++ {
		msg := audiolab.Message{
			SessionId: sessionId,
			Data: audiolab.Data{
				Internal: newMockIntBuffer(),
				External: newMockIntBuffer(),
			},
			IsClosed: false,
		}
		dispatchChan <- msg
	}
	msgClose := audiolab.Message{
		SessionId: sessionId,
		IsClosed:  true,
	}
	dispatchChan <- msgClose

	close(dispatchChan)
	fileStorage.Close()

	// THEN
	time.Sleep(500 * time.Millisecond)
	finalGoroutines := getGoroutineCount()
	// Autoriser un léger dépassement dû aux goroutines du runtime
	if finalGoroutines > initialGoroutines+1 {
		t.Errorf("Fuite de goroutines détectée. Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	} else {
		t.Logf("Aucune fuite de goroutine détectée. Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	}
}

func newMockIntBuffer() *externalaudio.IntBuffer {
	return &externalaudio.IntBuffer{
		Format: &externalaudio.Format{
			SampleRate:  8000,
			NumChannels: 1,
		},
		Data: []int{0, 1, 2, 3, 4, 5}, // Données mockées
	}
}

func initStorage() (chan audiolab.Message, *FileStorage, error) {
	// Créer un canal de dispatch
	dispatchChan := make(chan audiolab.Message, 100)

	// Initialiser le mock uploader
	mockUploader := utils.NewMockS3Uploader()

	// Initialiser FileStorage
	fileStorage, err := NewFileStorage(dispatchChan, mockUploader, utils.AppConfig{})
	return dispatchChan, fileStorage, err
}

// Helper pour récupérer le nombre actuel de goroutines
func getGoroutineCount() int {
	return runtime.NumGoroutine()
}

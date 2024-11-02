package outbound

import (
	"audio-server/audiolab"
	"audio-server/utils"
	"context"
	"fmt"
	"github.com/go-audio/wav"
	"go.uber.org/zap"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileStorage struct {
	mu           sync.Mutex
	wg           sync.WaitGroup
	dispatchChan <-chan audiolab.Message
	workers      map[string]chan audiolab.Message
	uploader     utils.FileUploader
	tempsDir     string
}

func NewFileStorage(dispatchChan <-chan audiolab.Message, uploader utils.FileUploader, appConfig utils.AppConfig) (*FileStorage, error) {
	tempsDir := appConfig.AudioTempRep
	if err := os.Mkdir(tempsDir, fs.ModeDir); err != nil {
		if os.IsExist(err) {
			utils.Logger.Debug("Le répertoire existe déjà",
				zap.String("tempsDir", tempsDir))
		} else {
			return nil, err
		}
	}
	if err := os.Chmod(tempsDir, fs.ModePerm); err != nil {
		return nil, err
	}

	fileStorage := &FileStorage{
		workers:      make(map[string]chan audiolab.Message),
		dispatchChan: dispatchChan,
		uploader:     uploader,
		tempsDir:     tempsDir,
	}
	go fileStorage.run()
	return fileStorage, nil
}

func (f *FileStorage) run() {
	utils.Logger.Info("Start FileStorage")
	for event := range f.dispatchChan {
		f.handleMessage(event)
	}
	f.wg.Wait()
}

func (f *FileStorage) handleMessage(s audiolab.Message) {
	f.mu.Lock()
	ch, exists := f.workers[s.SessionId]
	if !exists && !s.IsClosed {
		ch = make(chan audiolab.Message, 10)
		f.workers[s.SessionId] = ch
		f.wg.Add(1)
		go f.worker(s.SessionId, ch)
		utils.Logger.Debug("Goroutine créée", zap.String("SessionId", s.SessionId))
	}
	f.mu.Unlock()

	if exists {
		if s.IsClosed {
			// Envoyer le message de fermeture
			ch <- s
			// Fermer le canal
			close(ch)
			// Nettoyer le worker de la map
			f.mu.Lock()
			delete(f.workers, s.SessionId)
			f.mu.Unlock()
		} else {
			// Envoyer la structure au canal
			ch <- s
		}
	} else if !s.IsClosed {
		// Envoyer la structure au nouveau canal
		ch <- s
	}
}

func (f *FileStorage) worker(sessionId string, ch chan audiolab.Message) {
	defer f.wg.Done()
	packageNumber := 0
	// init file
	fileInternalName, fileInternal, err := f.createTempsFile(sessionId, "internal")
	if err != nil {
		utils.Logger.Error("Erreur lors de la création du fichier temporaire internal",
			zap.String("sessionId", sessionId),
			zap.String("fileName", fileInternalName),
			zap.Error(err))
		return
	}

	fileExternalName, fileExternal, err := f.createTempsFile(sessionId, "external")
	if err != nil {
		utils.Logger.Error("Erreur lors de la création du fichier temporaire external",
			zap.String("sessionId", sessionId),
			zap.String("fileName", fileInternalName),
			zap.Error(err))
		return
	}

	// Initialisation de l'encodeur WAV avec les paramètres PCM = 1
	encoderInternal := wav.NewEncoder(fileInternal, 8000, 16, 1, 1)
	encoderExternal := wav.NewEncoder(fileExternal, 8000, 16, 1, 1)

	for {
		select {
		case s, ok := <-ch:
			if !ok {
				// Canal fermé, terminer la goroutine
				utils.Logger.Debug("Session terminé",
					zap.String("SessionId", s.SessionId))
				return
			}

			utils.Logger.Debug("Reçu data",
				zap.String("SessionId", s.SessionId),
				zap.Bool("IsClosed", s.IsClosed))

			if s.IsClosed {
				utils.Logger.Debug("Fin de session, fermé FileStorage",
					zap.String("SessionId", s.SessionId))
				var errs []error
				// upload sur S3 et delete les temporary file
				// INTERNAL
				if err := encoderInternal.Close(); err != nil {
					utils.Logger.Error("Erreur lors de la fermeture du encoderInternal", zap.Error(err))
					errs = append(errs, err)
				}
				if err := f.uploadFile(fileInternal, fileInternalName); err != nil {
					errs = append(errs, err)
				}
				if err := fileInternal.Close(); err != nil {
					utils.Logger.Error("Erreur lors de la fermeture du fichier WAV", zap.Error(err))
					errs = append(errs, err)
				}
				if err := deleteLocalFile(fileInternal); err != nil {
					errs = append(errs, err)
				}
				// EXTERNAL
				if err := encoderExternal.Close(); err != nil {
					utils.Logger.Error("Erreur lors de la fermeture du encoderExternal", zap.Error(err))
					errs = append(errs, err)
				}
				if err := f.uploadFile(fileExternal, fileExternalName); err != nil {
					errs = append(errs, err)
				}
				if err := fileExternal.Close(); err != nil {
					utils.Logger.Error("Erreur lors de la fermeture du fichier WAV", zap.Error(err))
					errs = append(errs, err)
				}
				if err := deleteLocalFile(fileExternal); err != nil {
					errs = append(errs, err)
				}
				if len(errs) == 0 {
					utils.Logger.Info("Fichier WAV enregistré avec succès.")
					errs = append(errs, err)
				}
				return
			}

			packageNumber++
			// Écrire les échantillons PCM dans l'encodeur WAV
			if err := encoderInternal.Write(s.Data.Internal); err != nil {
				utils.Logger.Error("Erreur lors de l'écriture des données PCM dans Internal WAV",
					zap.Int("packageNumber", packageNumber),
					zap.Error(err))
			}

			// Écrire les échantillons PCM dans l'encodeur WAV
			if err := encoderExternal.Write(s.Data.External); err != nil {
				utils.Logger.Error("Erreur lors de l'écriture des données PCM dans External WAV",
					zap.Int("packageNumber", packageNumber),
					zap.Error(err))
			}

		}
	}
}

func (f *FileStorage) createTempsFile(sessionId, fileType string) (string, *os.File, error) {
	fileInternalName := fileName(fileType, sessionId)
	fileInternal, err := os.Create(filepath.Join(f.tempsDir, fileInternalName))
	return fileInternalName, fileInternal, err
}

func (f *FileStorage) Close() {

}

func (f *FileStorage) uploadFile(file *os.File, fileName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rep := time.Now().Format("20060102")
	err := f.uploader.UploadFile(ctx, file, rep+"/"+fileName)
	if err != nil {
		utils.Logger.Error("Erreur lors de la uploade file ",
			zap.String("fileName", fileName),
			zap.Int64("fileSize", fileSize(file.Name())),
			zap.Error(err))
	}
	return err
}

func deleteLocalFile(file *os.File) error {
	// Supprimer le fichier local
	err := os.Remove(file.Name())
	if err != nil {
		utils.Logger.Error("Erreur lors de la suppression du fichier local",
			zap.String("fileName", file.Name()),
			zap.Error(err))
	}
	return err
}

func fileName(source, sessionId string) string {
	return fmt.Sprintf("%s_audio_%s.wav", source, sessionId)
}

func fileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		utils.Logger.Error("Erreur fileSize",
			zap.String("fileName", filePath),
			zap.Error(err))
		return 0
	}
	return info.Size()
}

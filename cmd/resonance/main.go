package main

import (
	"fmt"

	"github.com/yadneshx17/resonance/internal/playback"
)

func main() {
	player := playback.NewPlayer()
	// player.Load(playback.Track{Path: "shiv2.mp3"})
	// player.Play()

	// time.Sleep(5 * time.Second)
	// player.Pause()
	// fmt.Println("[-] Paused")

	posi := player.Position()
	fmt.Printf("Posi: %v", posi.Seconds())

	// time.Sleep(5 * time.Second)
	// player.Resume()
	// fmt.Println("\n[-] Resume")

	// posi2 := player.Position()
	// fmt.Printf("Posi: %v", posi2.Seconds())

	// player.Seek(15) // gotta see how to work with this

	// time.Sleep(15 * time.Second)
	// player.Stop()

	// --------------------------------------------------------

	// Queue
	// queue := playback.NewQueue()
	// queue.Add(playback.Track{Path: "angel.mp3"})
	// queue.Add(playback.Track{Path: "angel2.mp3"})
	// queue.Add(playback.Track{Path: "shiv2.mp3"})

	// queue.PrintList()

	// track, _ := queue.Current()

	// player.Load(track)

	// player.Play()

	// time.Sleep(10 * time.Second)
	// player.Stop()
	// // player.Wait() // enable when playing complete song

	// fmt.Println("going into loop")
	// for {
	// 	track, ok := queue.Next()
	// 	if !ok {
	// 		fmt.Println("no more tracks")
	// 		break // no more tracks
	// 	}
	// 	player.Load(track)
	// 	player.Play()
	// 	time.Sleep(10 * time.Second)
	// 	player.Stop()
	// 	// player.Wait() // complete song
	// }

	// -------------------------------------------------------------
	// Scan Dir

	queue := playback.NewQueue()
	queue.PopulateQueue("Music")
	queue.PrintTracksInQueue()
}

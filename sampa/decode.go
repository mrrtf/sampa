package sampa

// var NumberOfProcessedEvents int = 0
// var elinks []Payload
// var gbt *bitset.BitSet
// var inData bool
// var nextCheckPoint int
//
// func init() {
// 	for i := 0; i < 40; i++ {
// 		elinks = append(elinks, Payload{BitSet: *(bitset.New(100000))})
// 	}
// 	gbt = bitset.New(80)
// 	inData = false
// 	nextCheckPoint = 0
// }

// func ProcessEvent(event EventType) {
//
// 	if !inData {
// 		elinks[0].Clear()
// 		nextCheckPoint = SyncPattern.Length()
// 	}
//
// 	fmt.Printf("elinks[0] (%d) at SOE %d inData %v nextCheckPoint %v\n", elinks[0].Length(), event.Header().EventID, inData, nextCheckPoint)
// 	fmt.Println(event)
//
// 	if !event.HasPayload() {
// 		fmt.Println("Event has no payload")
// 		return
// 	}
//
// 	_, err := event.SOP()
//
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	n := len(event.Data())
// 	// n := 700
//
// 	var lookingForSync = true
// 	var lookingForHeader = false
// 	var bs bitset.BitSet
// 	var clusters []Cluster
//
// 	nsync := 0
// 	for i := 0; i < n; i += 3 {
//
// 		previousLength := elinks[0].Length()
//
// 		err := GBTword(event.Data()[i:i+3], gbt)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
//
// 		err = dispatchGBTword(gbt, elinks)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
//
// 		l := elinks[0].Length()
//
// 		if l-previousLength != nBitsPerChannel {
// 			log.Fatal("SCARAMBOUILLE")
// 		}
//
// 		if l == nextCheckPoint {
//
// 			if lookingForSync || lookingForHeader {
// 				bs = *(elinks[0].Last(SyncPattern.Length()))
// 			} else if inData {
// 				fmt.Println("Would have to deal with ", l, " bits of data here", elinks[0].Length())
// 				Decode(clusters, &elinks[0])
// 				lookingForHeader = true
// 				inData = false
// 				elinks[0].Clear()
// 				nextCheckPoint = SyncPattern.Length()
// 				continue
// 			} else {
// 				log.Fatal("I'm lost")
// 			}
//
// 			sdh := ataHeader{BitSet: bs}
//
// 			if lookingForSync {
// 				if sdh.IsEqual(SyncPattern.BitSet) {
// 					if sdh.PKT() != uint8(SyncPKT) {
// 						log.Fatal("something's really wrong")
// 					}
// 					nsync++
// 					lookingForSync = false
// 					lookingForHeader = true
// 					elinks[0].Clear()
// 					nextCheckPoint = 50
// 				} else {
// 					nextCheckPoint += nBitsPerChannel
// 				}
// 			} else if lookingForHeader {
// 				// fmt.Print(" PKT=", sdh.PKT())
// 				if sdh.PKT() == uint8(SyncPKT) {
// 					// a sync again, still looking for header...
// 					if !sdh.IsEqual(SyncPattern.BitSet) {
// 						log.Fatal("PKT=2 but not a sync header ?")
// 					}
// 					nsync++
// 					elinks[0].Clear()
// 					nextCheckPoint = 50
// 				} else if sdh.PKT() == uint8(HeartBeatPKT) {
// 					log.Fatal("HEARTBEAT found. Implement some logic here")
// 				} else if sdh.PKT() == uint8(DataPKT) {
// 					fmt.Println(sdh.StringAnnotated("\n"))
// 					lookingForHeader = false
// 					dataToGo := sdh.NumWords()
// 					fmt.Println(dataToGo, " 10-bits words to read")
// 					elinks[0].Clear()
// 					nextCheckPoint = int(dataToGo * 10)
// 					inData = true
// 					continue
// 				}
// 			}
// 		}
// 	}
//
// 	fmt.Printf("\nelinks[0] (%d) at EOE %v = %v nsync=%v\n", elinks[0].Length(), event.Header().EventID, elinks[0].Length(), nsync)
//
// 	fmt.Println()
// }

package master

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/codeuniversity/al-master/metrics"
	"github.com/codeuniversity/al-master/websocket"
	"github.com/codeuniversity/al-proto"
	websocketConn "github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

const (
	statesFolderName = "states"
)

type MasterState interface {
	Sync(s *Server)
	SendStepDone(s *Server) (*proto.Empty, error)
	//OtherStepDone() bool
}

type ChiefMaster struct {
}

func (c ChiefMaster) Sync(s *Server) {
	for {
		_, err := c.SendStepDone(s)
		if err != nil {
			fmt.Println("failed to send step done signal to sub master", err)
		} else {
			break
		}
	}
	for {
		if s.SubMasterData.ChiefMasterStepDone {
			break
		}
	}
}

func (c ChiefMaster) SendStepDone(s *Server) (*proto.Empty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.SubMasterCommunicationClient.ChiefMasterStepDone(ctx, &proto.Empty{})
}

type SubMaster struct {
}

func (u SubMaster) Sync(s *Server) {
	for {
		_, err := u.SendStepDone(s)
		if err != nil {
			fmt.Println("failed to send step done signal to chief master", err)
		} else {
			break
		}
	}
	for {
		if s.ChiefMasterData.SubMasterStepDone {
			break
		}
	}
}

func (u SubMaster) SendStepDone(s *Server) (*proto.Empty, error) {
	if s.SubMasterData.ChiefMasterCommunicationClient == nil {
		client, err := createMasterCommunicationClient(s.ChiefMasterAddress)
		if err != nil {
			fmt.Println("couldn't create connection to chief master")
			return nil, err
		}
		s.ChiefMasterCommunicationClient = client
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.ChiefMasterCommunicationClient.SubMasterStepDone(ctx, &proto.Empty{})
}

//MasterMode contains the server's mode it is currently in
type MasterMode struct {
	StandAloneMaster bool
	ChiefMaster      bool
	SubMaster        bool
}

//ServerConfig contains config data for Server
type ServerConfig struct {
	ConnBufferSize    int
	GRPCPort          int
	HTTPPort          int
	StateFileName     string
	LoadLatestState   bool
	BigBangConfigPath string
	BucketWidth       int
	XAxisSpawnCenter  float32
	MasterMode
}

// TODO eventually: shutdown logic
// TODO eventually: make client compatible
// TODO eventually: clean code

//Server that manages cell changes
type Server struct {
	ServerConfig

	*SimulationState

	cisClientPool               *CISClientPool
	websocketConnectionsHandler *websocket.ConnectionsHandler

	grpcServer *grpc.Server
	httpServer *http.Server

	ChiefMasterData
	SubMasterData

	MasterState
}

//ChiefMasterData will hold additional data that is needed if the master is in ChiefMasterData mode
type ChiefMasterData struct {
	SubMasterCommunicationClient        proto.MasterCommunicationServiceClient
	SubMasterStepDone                   bool
	CellsOutsideSubRespReceivedThisStep bool
}

//SubMasterData will hold additional data that is needed if the master is in Chief SubMasterData mode
type SubMasterData struct {
	ChiefMasterCommunicationClient proto.MasterCommunicationServiceClient
	ChiefMasterAddress             string
	InitializedByChiefMaster       bool
	XAxisResponsibilityStart       float32
	ProximityCells                 []*proto.Cell
	//Shutdown                       bool
	ChiefMasterStepDone            bool
	ProximityCellsReceivedThisStep bool
}

//NewServer with given config
func NewServer(config ServerConfig, subMasterData SubMasterData) *Server {
	clientPool := NewCISClientPool(config.ConnBufferSize)

	return &Server{
		ServerConfig:                config,
		SubMasterData:               subMasterData,
		websocketConnectionsHandler: websocket.NewConnectionsHandler(),
		cisClientPool:               clientPool,
	}
}

//Init loads state from a file or by asking a cis instance for a new BigBang depending on ServerConfig
func (s *Server) Init() {
	s.initPrometheus()
	go s.listen()

	if s.SubMaster {
		_, err := s.registerToChiefMaster()
		if err != nil {
			fmt.Println("Failed to register at chief master", err)
			panic(err)
		} else {
			fmt.Println("Successfully registered at chief master")
		}
		return
	}

	if s.StandAloneMaster {
		fmt.Println("Waiting for sub master")
		s.waitForSubMasterRegistration()
	}

	if s.StateFileName != "" {
		simulationState, err := LoadSimulationState(filepath.Join(statesFolderName, s.StateFileName))
		if err != nil {
			fmt.Println("\nLoading state from filepath failed, exiting now", err)
			panic(err)
		}
		s.SimulationState = simulationState
		return
	}

	//TODO: make "load latest state" work with sub master, only fetchBigBang considered for now
	if s.LoadLatestState {
		simulationState, err := LoadLatestSimulationState()
		if err != nil {
			fmt.Println("\nLoading latest state failed, exiting now", err)
			panic(err)
		}
		s.SimulationState = simulationState
		return
	}
	s.fetchBigBang()

	_, err := s.initializeSubMasterClient()
	if err != nil {
		fmt.Println("failed to initialize sub master")
	} else {
		fmt.Println("initialized sub master")
	}
}

//Run offloads the computation of changes to cis
func (s *Server) Run() {
	if s.MasterMode.SubMaster {
		fmt.Println("Waiting to get initialized by chief master")
		for {
			if s.InitializedByChiefMaster {
				break
			}
		}
		fmt.Println("got initialized by chief master")
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		if s.SubMaster{ s.MasterState = SubMaster{}}
		if s.ChiefMaster{s.MasterState = ChiefMaster{}}

		if len(s.CellBuckets.AllCells()) == 0 {
			fmt.Println("no cells remaining, stopping...")
			break
		}

		if len(signals) != 0 {
			fmt.Println("Received Signal:", <-signals)
			break
		}

		/*if s.SubMasterData.Shutdown {
			break
		}*/

		s.step()


		s.MasterState.Sync(s)

		if s.SubMaster {

			cellsOutsideResponsibility, cellsInsideResponsibility := s.checkForCellResponsibility()

			_, err := s.returnCellsToChiefMaster(cellsOutsideResponsibility)
			if err != nil {
				fmt.Println("failed to return outside responsibility cells to chief master")
				panic(err)
			} else {
				s.CellBuckets = CreateBuckets(cellsInsideResponsibility, uint(s.BucketWidth))
			}

			for {
				if s.SubMasterData.ProximityCellsReceivedThisStep {
					break
				}
			}

			s.ProximityCellsReceivedThisStep = false
			s.SubMasterData.ChiefMasterStepDone = false
		}

		if s.ChiefMaster {

			_, err := s.sendSubMasterCellsInProximity()
			if err != nil {
				fmt.Println("failed to send sub master cells in proximity")
				panic(err)
			}

			for {
				if s.ChiefMasterData.CellsOutsideSubRespReceivedThisStep {
					break
				}
			}

			s.CellsOutsideSubRespReceivedThisStep = false
			s.ChiefMasterData.SubMasterStepDone = false
		}
	}
	s.shutdown()
}

//Register cis-slave and create clients to make the slave useful
func (s *Server) Register(ctx context.Context, registration *proto.SlaveRegistration) (*proto.Empty, error) {
	for i := 0; i < int(registration.Threads); i++ {
		client, err := createCellInteractionClient(registration.Address)
		if err != nil {
			return nil, err
		}
		s.cisClientPool.AddClient(client)
		metrics.CISClientCount.Inc()
	}
	fmt.Println("CIS registered")
	return &proto.Empty{}, nil
}

//RegisterSubMaster registers sub master and creates client to make the sub master useful
func (s *Server) RegisterSubMaster(ctx context.Context, registration *proto.SubMasterRegistration) (*proto.Empty, error) {
	fmt.Println("got request to register sub master")
	masterCommunicationClient, err := createMasterCommunicationClient(registration.Address)
	if err != nil {
		return nil, err
	}
	s.ChiefMasterData.SubMasterCommunicationClient = masterCommunicationClient
	s.ChiefMaster = true
	return &proto.Empty{}, err
}

/*//UnregisterSubMaster unregisters sub master from chief master
func (s *Server) UnregisterSubMaster(ctx context.Context, transfer *proto.CellTransfer) (*proto.Empty, error) {
	s.CellBuckets.Merge(CreateBuckets(transfer.CellsToTransfer, uint(s.BucketWidth)))
	s.ChiefMaster = false
	s.StandAloneMaster = true
	s.SubMasterCommunicationClient = nil
	return &proto.Empty{}, nil
}*/

func (s *Server) InitializeSubMaster(ctx context.Context, initializationData *proto.SubMasterInitialization) (*proto.Empty, error) {
	s.BucketWidth = int(initializationData.BucketWidth)
	buckets := CreateBuckets(initializationData.CellsToTransfer, uint(s.BucketWidth))
	s.SimulationState = NewSimulationState(buckets)
	s.ProximityCells = initializationData.CellsInProximity
	s.XAxisResponsibilityStart = initializationData.XAxisResponsibilityStart
	s.InitializedByChiefMaster = true
	return &proto.Empty{}, nil
}

/*func (s *Server) ShutDownSubMaster(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	_, err := s.returnCellsToChiefMaster(s.CellBuckets.AllCells())
	if err != nil {
		fmt.Println("failed to send back cells to chief master")
		return nil, err
	}
	s.Shutdown = true
	return nil, err
}*/

func (s *Server) TransferCellsOutsideResponsibility(ctx context.Context, transfer *proto.CellTransfer) (*proto.Empty, error) {
	s.CellBuckets.Merge(CreateBuckets(transfer.CellsToTransfer, uint(s.BucketWidth)))
	s.CellsOutsideSubRespReceivedThisStep = true
	return &proto.Empty{}, nil
}

func (s *Server) UpdateCellsInProximity(ctx context.Context, transfer *proto.CellTransfer) (*proto.Empty, error) {
	s.ProximityCells = transfer.CellsToTransfer
	s.ProximityCellsReceivedThisStep = true
	return &proto.Empty{}, nil
}

func (s *Server) SubMasterStepDone(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	s.SubMasterData.ChiefMasterStepDone = true
	return &proto.Empty{}, nil
}

func (s *Server) ChiefMasterStepDone(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	s.ChiefMasterData.SubMasterStepDone = true
	return &proto.Empty{}, nil
}

func (s *Server) registerToChiefMaster() (*proto.Empty, error) {
	conn, err := grpc.Dial(s.ChiefMasterAddress, grpc.WithInsecure())

	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("closing chief master connection failed")
		}
	}()

	if err != nil {
		fmt.Println("couldn't dial to chief master: ", err)
		return &proto.Empty{}, err
	}
	chiefMasterClient := proto.NewSubMasterRegistrationServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := chiefMasterClient.RegisterSubMaster(ctx, &proto.SubMasterRegistration{Address: fmt.Sprintf("localhost:%d", s.GRPCPort)})
	return resp, err
}

/*func (s *Server) unregisterFromChiefMaster() (*proto.Empty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cellTransfer := &proto.CellTransfer{
		CellsToTransfer: s.CellBuckets.AllCells(),
	}
	resp, err := s.ChiefMasterData.SubMasterCommunicationClient.UnregisterSubMaster(ctx, cellTransfer)
	if err != nil {
		fmt.Println("UnregisterSubMaster call failed", err)
	}
	return resp, err
}*/

func (s *Server) waitForSubMasterRegistration() {
	for {
		if s.ChiefMaster {
			fmt.Println("\nSub Master successfully registered")
			break
		}
	}
}

func (s *Server) initializeSubMasterClient() (*proto.Empty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var chiefMasterCells []*proto.Cell
	var subMasterCells []*proto.Cell
	var cellsInSubProximity []*proto.Cell

	for _, cell := range s.SimulationState.CellBuckets.AllCells() {
		if cell.Pos.X < s.XAxisSpawnCenter {
			if cell.Pos.X > (s.XAxisSpawnCenter - float32(s.BucketWidth)) {
				cellsInSubProximity = append(cellsInSubProximity, cell)
			}
			chiefMasterCells = append(chiefMasterCells, cell)
		} else {
			subMasterCells = append(subMasterCells, cell)
		}
	}
	s.SimulationState.CellBuckets = CreateBuckets(chiefMasterCells, uint(s.BucketWidth))

	subMasterInitialization := &proto.SubMasterInitialization{
		CellsToTransfer:          subMasterCells,
		CellsInProximity:         cellsInSubProximity,
		XAxisResponsibilityStart: s.XAxisSpawnCenter,
		BucketWidth:              int32(s.BucketWidth),
	}
	resp, err := s.ChiefMasterData.SubMasterCommunicationClient.InitializeSubMaster(ctx, subMasterInitialization)
	return resp, err
}

func (s *Server) returnCellsToChiefMaster(cells []*proto.Cell) (*proto.Empty, error) {
	if s.SubMasterData.ChiefMasterCommunicationClient == nil {
		client, err := createMasterCommunicationClient(s.ChiefMasterAddress)
		if err != nil {
			fmt.Println("couldn't create connection to chief master")
			return nil, err
		}
		s.ChiefMasterCommunicationClient = client
	}

	cellTransfer := &proto.CellTransfer{
		CellsToTransfer: cells,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := s.SubMasterData.ChiefMasterCommunicationClient.TransferCellsOutsideResponsibility(ctx, cellTransfer)
	return resp, err
}

func (s *Server) sendSubMasterCellsInProximity() (*proto.Empty, error) {
	cellsInSubProximity, err := s.cellsInProximityForSubMaster()
	if err != nil {
		return nil, err
	}
	proximityCellsTransfer := &proto.CellTransfer{
		CellsToTransfer: cellsInSubProximity,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := s.ChiefMasterData.SubMasterCommunicationClient.UpdateCellsInProximity(ctx, proximityCellsTransfer)
	return resp, err
}

func (s *Server) checkForCellResponsibility() (cellsOutsideResp []*proto.Cell, cellsInsideResp []*proto.Cell) {
	var noResponsibilityCells []*proto.Cell
	var responsibilityCells []*proto.Cell

	for _, cell := range s.CellBuckets.AllCells() {
		if cell.Pos.X < s.SubMasterData.XAxisResponsibilityStart {
			noResponsibilityCells = append(noResponsibilityCells, cell)
		} else {
			responsibilityCells = append(responsibilityCells, cell)
		}
	}
	return noResponsibilityCells, responsibilityCells
}

/*func (s *Server) signalSubMasterToShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := s.SubMasterCommunicationClient.ShutDownSubMaster(ctx, &proto.Empty{})
	return err
}*/

func (s *Server) cellsInProximityForSubMaster() ([]*proto.Cell, error) {
	var cellsInSubProximity []*proto.Cell

	for _, cell := range s.CellBuckets.AllCells() {
		if cell.Pos.X < s.XAxisSpawnCenter && cell.Pos.X < (s.XAxisSpawnCenter-float32(s.BucketWidth)) {
			cellsInSubProximity = append(cellsInSubProximity, cell)
		}
	}
	return cellsInSubProximity, nil
}

func (s *Server) initPrometheus() {
	prometheus.MustRegister(metrics.AmountOfBuckets)
	prometheus.MustRegister(metrics.AverageCellsPerBucket)
	prometheus.MustRegister(metrics.MedianCellsPerBucket)
	prometheus.MustRegister(metrics.MinCellsInBuckets)
	prometheus.MustRegister(metrics.MaxCellsInBuckets)
	prometheus.MustRegister(metrics.CISCallCounter)
	prometheus.MustRegister(metrics.CisCallDurationSeconds)
	prometheus.MustRegister(metrics.CISClientCount)
	prometheus.MustRegister(metrics.WebSocketConnectionsCount)

	http.Handle("/metrics", promhttp.Handler())
}

func (s *Server) shutdown() {
	s.closeConnections()

	/*if s.SubMaster {
		for {
			_, err := s.unregisterFromChiefMaster()
			if err != nil {
				fmt.Println("Couldn't transfer cells back to chief master before shutdown")
			} else {
				break
			}
		}
		return
	}*/

	/*if s.ChiefMaster {
		for {
			err := s.signalSubMasterToShutdown()
			if err != nil {
				fmt.Println("Couldn't send shutdown call to sub master")
			} else {
				break
			}
		}
	}*/

	err := s.saveState()
	if err == nil {
		fmt.Println("\nState successfully saved")
	} else {
		fmt.Println("\nState could not be saved:", err)
	}
}

func (s *Server) closeConnections() {
	s.websocketConnectionsHandler.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		fmt.Println("Couldn't shutdown http server", err)
	}
	s.grpcServer.Stop()
}

func (s *Server) fetchBigBang() {
	config, err := BigBangConfigFromPath(s.ServerConfig.BigBangConfigPath)
	if err != nil {
		panic(err)
	}

	s.ServerConfig.XAxisSpawnCenter = config.End.X / 2.0

	fmt.Println("Waiting for CIS to register")
	c := s.cisClientPool.GetClient()
	defer s.cisClientPool.AddClient(c)
	withTimeout(100*time.Second, func(ctx context.Context) {
		stream, err := c.BigBang(ctx, config.ToProto())
		if err != nil {
			panic(err)
		}
		cells := make([]*proto.Cell, 0, config.CellAmount)
		for {
			cell, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
				break
			}
			cells = append(cells, cell)
		}
		buckets := CreateBuckets(cells, uint(s.BucketWidth))
		s.SimulationState = NewSimulationState(buckets)
	})
}

func (s *Server) listen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", s.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s.grpcServer = grpc.NewServer()
	proto.RegisterSlaveRegistrationServiceServer(s.grpcServer, s)
	proto.RegisterMasterCommunicationServiceServer(s.grpcServer, s)

	if s.StandAloneMaster {
		proto.RegisterSubMasterRegistrationServiceServer(s.grpcServer, s)
	}

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	http.HandleFunc("/", s.websocketHandler)
	s.httpServer = &http.Server{Addr: fmt.Sprintf(":%v", s.HTTPPort), Handler: nil}
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

var upgrader = websocketConn.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	s.websocketConnectionsHandler.AddConnection(conn)
}

func (s *Server) broadcastCurrentState() {
	s.websocketConnectionsHandler.BroadcastCells(s.CellBuckets.AllCells())
}

func (s *Server) step() {
	UpdateBucketsMetrics(s.CellBuckets)
	doneChan := make(chan struct{})

	go s.processReturnedBatches(s.CurrentReturnedBatchChan(), doneChan)
	for key, bucket := range s.CellBuckets {
		if s.RequestInflightFromLastStep(key) {
			continue
		}

		var surroundingCells []*proto.Cell
		for _, otherKey := range key.SurroundingKeys(s.BucketWidth) {
			if otherBucket, ok := s.CellBuckets[otherKey]; ok {
				surroundingCells = append(surroundingCells, otherBucket...)
			}
		}
		s.CurrentWaitGroup().Add(1)
		batch := &proto.CellComputeBatch{
			CellsToCompute:   bucket,
			CellsInProximity: surroundingCells,
			TimeStep:         s.TimeStep,
			BatchKey:         string(key),
		}
		go s.callCIS(batch, s.CurrentWaitGroup(), s.CurrentReturnedBatchChan())
	}

	s.CurrentWaitGroup().Wait()
	s.Cycle()
	<-doneChan
	s.TimeStep++
	fmt.Println(s.TimeStep, ": ", len(s.CellBuckets.AllCells()))
	s.broadcastCurrentState()
}

func (s *Server) callCIS(batch *proto.CellComputeBatch, wg *sync.WaitGroup, returnedBatchChan chan *proto.CellComputeBatch) {
	metrics.CISCallCounter.Inc()
	looping := true
	for looping {
		c := s.cisClientPool.GetClient()
		withTimeout(10*time.Second, func(ctx context.Context) {
			start := time.Now()
			returnedBatch, err := c.ComputeCellInteractions(ctx, batch)
			metrics.CisCallDurationSeconds.Observe(time.Since(start).Seconds())
			if err == nil {
				s.cisClientPool.AddClient(c)
				returnedBatchChan <- returnedBatch
				looping = false
			} else {
				metrics.CISClientCount.Dec()
			}
		})
	}
	wg.Done()
}

func (s *Server) processReturnedBatches(returnedBatchChan chan *proto.CellComputeBatch, doneChan chan struct{}) {
	nextBuckets := Buckets{}
	doneNeighbourBuckets := map[BucketKey]int{}

	for returnedBatch := range returnedBatchChan {
		returnedBuckets := CreateBuckets(returnedBatch.CellsToCompute, uint(s.BucketWidth))
		nextBuckets.Merge(returnedBuckets)
		bucketKey := BucketKey(returnedBatch.BatchKey)

		// call cis for next step if possible
		keysToCheck := append(bucketKey.SurroundingKeys(s.BucketWidth), bucketKey)

		for _, key := range keysToCheck {
			doneNeighbourBuckets[key]++
			bucket, exists := nextBuckets[key]
			if !exists {
				continue
			}
			if len(bucket) == 0 || doneNeighbourBuckets[key] < 27 || s.RequestInflight(key) {
				continue
			}

			var surroundingCells []*proto.Cell
			for _, surroundingKey := range key.SurroundingKeys(s.BucketWidth) {
				surroundingBucket := nextBuckets[surroundingKey]
				surroundingCells = append(surroundingCells, surroundingBucket...)
			}

			batch := &proto.CellComputeBatch{
				CellsToCompute:   bucket,
				CellsInProximity: surroundingCells,
				TimeStep:         s.TimeStep + 1,
				BatchKey:         string(key),
			}
			s.NextWaitGroup().Add(1)
			go s.callCIS(batch, s.NextWaitGroup(), s.NextReturnedBatchChan())

			s.MarkRequestInflight(key)
		}
	}
	s.CellBuckets = nextBuckets
	doneChan <- struct{}{}
}

func withTimeout(timeout time.Duration, f func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	f(ctx)
}

func createCellInteractionClient(address string) (proto.CellInteractionServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewCellInteractionServiceClient(conn), nil
}

func createMasterCommunicationClient(address string) (proto.MasterCommunicationServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewMasterCommunicationServiceClient(conn), nil
}

package master

import "net"
import "net/rpc"
import "log"
import "time"
import "fmt"

import "main/lib/process"

// RemoteMaster is a struct that holds the master instance.
type RemoteMaster struct {
	master *Master // Master instance
}

// RemoteClient is a struct that holds the remote client instance.
type RemoteClient struct {
	conn *rpc.Client // RpcConnection for the remote client.
}

// GoBin is a struct that represents the necessary arguments for a go binary to be built.
type GoBin struct {
	SourcePath string   // SourcePath is the package path. (Ex: github.com/topfreegames/apm)
	Name       string   // Name is the process name that will be given to the process.
	KeepAlive  bool     // KeepAlive will determine whether APM should keep the proc live or not.
	Args       []string // Args is an array containing all the extra args that will be passed to the binary after compilation.
}

type ProcDataResponse struct {
	Name string
	Pid int
	Status *process.ProcStatus
	KeepAlive bool
}

type ProcResponse struct {
	Procs []*ProcDataResponse
}
// Save will save the current running and stopped processes onto a file.
// Returns an error in case there's any.
func (remote_master *RemoteMaster) Save(req string, ack *bool) error {
	req = ""
	*ack = true
	return remote_master.master.SaveProcs()
}

// Resurrect will restore all previously save processes.
// Returns an error in case there's any.
func (remote_master *RemoteMaster) Resurrect(req string, ack *bool) error {
	req = ""
	*ack = true
	return remote_master.master.Revive()
}

// StartGoBin will build a binary based on the arguments passed on goBin, then it will start the process
// and keep it alive if KeepAlive is set to true.
// It returns an error and binds true to ack pointer.
func (remote_master *RemoteMaster) StartGoBin(goBin *GoBin, ack *bool) error {
	preparable, output, err := remote_master.master.Prepare(goBin.SourcePath, goBin.Name, "go", goBin.KeepAlive, goBin.Args)
	*ack = true
	if err != nil {
		return fmt.Errorf("ERROR: %s OUTPUT: %s", err, string(output))
	}
	return remote_master.master.RunPreparable(preparable)
}

// RestartProcess will restart a process that was previously built using GoBin.
// It returns an error in case there's any.
func (remote_master *RemoteMaster) RestartProcess(procName string, ack *bool) error {
	*ack = true
	return remote_master.master.RestartProcess(procName)
}

// StartProcess will start a process that was previously built using GoBin.
// It returns an error in case there's any.
func (remote_master *RemoteMaster) StartProcess(procName string, ack *bool) error {
	*ack = true
	return remote_master.master.StartProcess(procName)
}

// StopProcess will stop a process that is currently running.
// It returns an error in case there's any.
func (remote_master *RemoteMaster) StopProcess(procName string, ack *bool) error {
	*ack = true
	return remote_master.master.StopProcess(procName)
}

// MonitStatus will query for the status of each process and bind it to procs pointer list.
// It returns an error in case there's any.
func (remote_master *RemoteMaster) MonitStatus(req string, response *ProcResponse) error {
	req = ""
	procs := remote_master.master.ListProcs()
	procsResponse := []*ProcDataResponse{}
	for id := range procs {
		proc := procs[id]
		procData := &ProcDataResponse {
			Name: proc.Identifier(),
			Pid: proc.GetPid(),
			Status: proc.GetStatus(),
			KeepAlive: proc.ShouldKeepAlive(),
		}
		procsResponse = append(procsResponse, procData)
	}
	*response = ProcResponse {
		Procs: procsResponse,
	}
	return nil
}

// DeleteProcess will delete a process with name procName.
// It returns an error in case there's any.
func (remote_master *RemoteMaster) DeleteProcess(procName string, ack *bool) error {
	*ack = true
	return remote_master.master.DeleteProcess(procName)
}

// Stop will stop APM remote server.
// It returns an error in case there's any.
func (remote_master *RemoteMaster) Stop() error {
	return remote_master.master.Stop()
}

// StartRemoteMasterServer starts a remote APM server listening on dsn address and binding to
// configFile.
// It returns a RemoteMaster instance.
func StartRemoteMasterServer(dsn string, configFile string) *RemoteMaster {
	remoteMaster := &RemoteMaster{
		master: InitMaster(configFile),
	}
	rpc.Register(remoteMaster)
	l, e := net.Listen("tcp", dsn)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	go rpc.Accept(l)
	return remoteMaster
}

// StartRemoteClient will start a remote client that can talk to a remote server that
// is already running on dsn address.
// It returns an error in case there's any or it could not connect within the timeout.
func StartRemoteClient(dsn string, timeout time.Duration) (*RemoteClient, error) {
	conn, err := net.DialTimeout("tcp", dsn, timeout)
	if err != nil {
		return nil, err
	}
	return &RemoteClient{conn: rpc.NewClient(conn)}, nil
}

// Save will save a list of procs onto a file.
// Returns an error in case there's any.
func (client *RemoteClient) Save() error {
	var started bool
	return client.conn.Call("RemoteMaster.Save", "", &started)
}

// Resurrect will restore all previously save processes.
// Returns an error in case there's any.
func (client *RemoteClient) Resurrect() error {
	var started bool
	return client.conn.Call("RemoteMaster.Resurrect", "", &started)
}

// StartGoBin is a wrapper that calls the remote StartsGoBin.
// It returns an error in case there's any.
func (client *RemoteClient) StartGoBin(sourcePath string, name string, keepAlive bool, args []string) error {
	goBin := &GoBin{
		SourcePath: sourcePath,
		Name:       name,
		KeepAlive:  keepAlive,
		Args:       args,
	}
	var started bool
	return client.conn.Call("RemoteMaster.StartGoBin", goBin, &started)
}

// RestartProcess is a wrapper that calls the remote RestartProcess.
// It returns an error in case there's any.
func (client *RemoteClient) RestartProcess(procName string) error {
	var started bool
	return client.conn.Call("RemoteMaster.RestartProcess", procName, &started)
}

// StartProcess is a wrapper that calls the remote StartProcess.
// It returns an error in case there's any.
func (client *RemoteClient) StartProcess(procName string) error {
	var started bool
	return client.conn.Call("RemoteMaster.StartProcess", procName, &started)
}

// StopProcess is a wrapper that calls the remote StopProcess.
// It returns an error in case there's any.
func (client *RemoteClient) StopProcess(procName string) error {
	var stopped bool
	return client.conn.Call("RemoteMaster.StopProcess", procName, &stopped)
}

// DeleteProcess is a wrapper that calls the remote DeleteProcess.
// It returns an error in case there's any.
func (client *RemoteClient) DeleteProcess(procName string) error {
	var deleted bool
	return client.conn.Call("RemoteMaster.DeleteProcess", procName, &deleted)
}

// MonitStatus is a wrapper that calls the remote MonitStatus.
// It returns a tuple with a list of process and an error in case there's any.
func (client *RemoteClient) MonitStatus() (ProcResponse, error) {
	var response *ProcResponse
	err := client.conn.Call("RemoteMaster.MonitStatus", "", &response)
	return *response, err
}

/*
 * nomad
 *
 * <h1 id=\"http-api\">HTTP API</h1> <p>The main interface to Nomad is a RESTful HTTP API. The API can query the current     state of the system as well as modify the state of the system. The Nomad CLI     actually invokes Nomad&#39;s HTTP for many commands.</p> <h2 id=\"version-prefix\">Version Prefix</h2> <p>All API routes are prefixed with <code>/v1/</code>.</p> <h2 id=\"addressing-and-ports\">Addressing and Ports</h2> <p>Nomad binds to a specific set of addresses and ports. The HTTP API is served via     the <code>http</code> address and port. This <code>address:port</code> must be accessible locally. If     you bind to <code>127.0.0.1:4646</code>, the API is only available <em>from that host</em>. If you     bind to a private internal IP, the API will be available from within that     network. If you bind to a public IP, the API will be available from the public     Internet (not recommended).</p> <p>The default port for the Nomad HTTP API is <code>4646</code>. This can be overridden via     the Nomad configuration block. Here is an example curl request to query a Nomad     server with the default configuration:</p> <pre><code class=\"language-shell-session\">$ curl http://127.0.0.1:4646/v1/agent/members </code></pre> <p>The conventions used in the API documentation do not list a port and use the     standard URL <code>localhost:4646</code>. Be sure to replace this with your Nomad agent URL     when using the examples.</p> <h2 id=\"data-model-and-layout\">Data Model and Layout</h2> <p>There are five primary nouns in Nomad:</p> <ul>     <li>jobs</li>     <li>nodes</li>     <li>allocations</li>     <li>deployments</li>     <li>evaluations</li> </ul> <p><a href=\"/img/nomad-data-model.png\"><img src=\"/img/nomad-data-model.png\" alt=\"Nomad Data Model\"></a></p> <p>Jobs are submitted by users and represent a <em>desired state</em>. A job is a     declarative description of tasks to run which are bounded by constraints and     require resources. Jobs can also have affinities which are used to express placement     preferences. Nodes are the servers in the clusters that tasks can be     scheduled on. The mapping of tasks in a job to nodes is done using allocations.     An allocation is used to declare that a set of tasks in a job should be run on a     particular node. Scheduling is the process of determining the appropriate     allocations and is done as part of an evaluation. Deployments are objects to     track a rolling update of allocations between two versions of a job.</p> <p>The API is modeled closely on the underlying data model. Use the links to the     left for documentation about specific endpoints. There are also &quot;Agent&quot; APIs     which interact with a specific agent and not the broader cluster used for     administration.</p> <h2 id=\"acls\">ACLs</h2> <p>Several endpoints in Nomad use or require ACL tokens to operate. The token are used to authenticate the request and determine if the request is allowed based on the associated authorizations. Tokens are specified per-request by using the <code>X-Nomad-Token</code> request header set to the <code>SecretID</code> of an ACL Token.</p> <p>For more details about ACLs, please see the <a href=\"https://learn.hashicorp.com/collections/nomad/access-control\">ACL Guide</a>.</p> <h2 id=\"authentication\">Authentication</h2> <p>When ACLs are enabled, a Nomad token should be provided to API requests using the <code>X-Nomad-Token</code> header. When using authentication, clients should communicate via TLS.</p> <p>Here is an example using curl:</p> <pre><code class=\"language-shell-session\">$ curl \\     --header &quot;X-Nomad-Token: aa534e09-6a07-0a45-2295-a7f77063d429&quot; \\     https://localhost:4646/v1/jobs </code></pre> <h2 id=\"namespaces\">Namespaces</h2> <p>Nomad has support for namespaces, which allow jobs and their associated objects     to be segmented from each other and other users of the cluster. When using     non-default namespace, the API request must pass the target namespace as an API     query parameter. Prior to Nomad 1.0 namespaces were Enterprise-only.</p> <p>Here is an example using curl:</p> <pre><code class=\"language-shell-session\">$ curl \\     https://localhost:4646/v1/jobs?namespace=qa </code></pre> <h2 id=\"blocking-queries\">Blocking Queries</h2> <p>Many endpoints in Nomad support a feature known as &quot;blocking queries&quot;. A     blocking query is used to wait for a potential change using long polling. Not     all endpoints support blocking, but each endpoint uniquely documents its support     for blocking queries in the documentation.</p> <p>Endpoints that support blocking queries return an HTTP header named     <code>X-Nomad-Index</code>. This is a unique identifier representing the current state of     the requested resource. On a new Nomad cluster the value of this index starts at 1. </p> <p>On subsequent requests for this resource, the client can set the <code>index</code> query     string parameter to the value of <code>X-Nomad-Index</code>, indicating that the client     wishes to wait for any changes subsequent to that index.</p> <p>When this is provided, the HTTP request will &quot;hang&quot; until a change in the system     occurs, or the maximum timeout is reached. A critical note is that the return of     a blocking request is <strong>no guarantee</strong> of a change. It is possible that the     timeout was reached or that there was an idempotent write that does not affect     the result of the query.</p> <p>In addition to <code>index</code>, endpoints that support blocking will also honor a <code>wait</code>     parameter specifying a maximum duration for the blocking request. This is     limited to 10 minutes. If not set, the wait time defaults to 5 minutes. This     value can be specified in the form of &quot;10s&quot; or &quot;5m&quot; (i.e., 10 seconds or 5     minutes, respectively). A small random amount of additional wait time is added     to the supplied maximum <code>wait</code> time to spread out the wake up time of any     concurrent requests. This adds up to <code>wait / 16</code> additional time to the maximum     duration.</p> <h2 id=\"consistency-modes\">Consistency Modes</h2> <p>Most of the read query endpoints support multiple levels of consistency. Since     no policy will suit all clients&#39; needs, these consistency modes allow the user     to have the ultimate say in how to balance the trade-offs inherent in a     distributed system.</p> <p>The two read modes are:</p> <ul>     <li>         <p><code>default</code> - If not specified, the default is strongly consistent in almost all             cases. However, there is a small window in which a new leader may be elected             during which the old leader may service stale values. The trade-off is fast             reads but potentially stale values. The condition resulting in stale reads is             hard to trigger, and most clients should not need to worry about this case.             Also, note that this race condition only applies to reads, not writes.</p>     </li>     <li>         <p><code>stale</code> - This mode allows any server to service the read regardless of             whether it is the leader. This means reads can be arbitrarily stale; however,             results are generally consistent to within 50 milliseconds of the leader. The             trade-off is very fast and scalable reads with a higher likelihood of stale             values. Since this mode allows reads without a leader, a cluster that is             unavailable will still be able to respond to queries.</p>     </li> </ul> <p>To switch these modes, use the <code>stale</code> query parameter on requests.</p> <p>To support bounding the acceptable staleness of data, responses provide the     <code>X-Nomad-LastContact</code> header containing the time in milliseconds that a server     was last contacted by the leader node. The <code>X-Nomad-KnownLeader</code> header also     indicates if there is a known leader. These can be used by clients to gauge the     staleness of a result and take appropriate action. </p> <h2 id=\"cross-region-requests\">Cross-Region Requests</h2> <p>By default, any request to the HTTP API will default to the region on which the     machine is servicing the request. If the agent runs in &quot;region1&quot;, the request     will query the region &quot;region1&quot;. A target region can be explicitly request using     the <code>?region</code> query parameter. The request will be transparently forwarded and     serviced by a server in the requested region.</p> <h2 id=\"compressed-responses\">Compressed Responses</h2> <p>The HTTP API will gzip the response if the HTTP request denotes that the client     accepts gzip compression. This is achieved by passing the accept encoding:</p> <pre><code class=\"language-shell-session\">$ curl \\     --header &quot;Accept-Encoding: gzip&quot; \\     https://localhost:4646/v1/... </code></pre> <h2 id=\"formatted-json-output\">Formatted JSON Output</h2> <p>By default, the output of all HTTP API requests is minimized JSON. If the client     passes <code>pretty</code> on the query string, formatted JSON will be returned.</p> <p>In general, clients should prefer a client-side parser like <code>jq</code> instead of     server-formatted data. Asking the server to format the data takes away     processing cycles from more important tasks.</p> <pre><code class=\"language-shell-session\">$ curl https://localhost:4646/v1/page?pretty </code></pre> <h2 id=\"http-methods\">HTTP Methods</h2> <p>Nomad&#39;s API aims to be RESTful, although there are some exceptions. The API     responds to the standard HTTP verbs GET, PUT, and DELETE. Each API method will     clearly document the verb(s) it responds to and the generated response. The same     path with different verbs may trigger different behavior. For example:</p> <pre><code class=\"language-text\">PUT /v1/jobs GET /v1/jobs </code></pre> <p>Even though these share a path, the <code>PUT</code> operation creates a new job whereas     the <code>GET</code> operation reads all jobs.</p> <h2 id=\"http-response-codes\">HTTP Response Codes</h2> <p>Individual API&#39;s will contain further documentation in the case that more     specific response codes are returned but all clients should handle the following:</p> <ul>     <li>200 and 204 as success codes.</li>     <li>400 indicates a validation failure and if a parameter is modified in the         request, it could potentially succeed.</li>     <li>403 marks that the client isn&#39;t authenticated for the request.</li>     <li>404 indicates an unknown resource.</li>     <li>5xx means that the client should not expect the request to succeed if retried.</li> </ul>
 *
 * API version: 1.1.0
 * Contact: support@hashicorp.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package testclient

import (
	"encoding/json"
)

// TaskEvent TaskEvent is an event that effects the state of a task and contains meta-data appropriate to the events type.
type TaskEvent struct {
	// Details is a map with annotated info about the event
	Details *map[string]string `json:"Details,omitempty"`
	// The maximum allowed task disk size. Deprecated, use Details[\"disk_limit\"] to access this.
	DiskLimit *int64 `json:"DiskLimit,omitempty"`
	// DisplayMessage is a human friendly message about the event
	DisplayMessage *string `json:"DisplayMessage,omitempty"`
	// Artifact Download fields Deprecated, use Details[\"download_error\"] to access this.
	DownloadError *string `json:"DownloadError,omitempty"`
	// Driver Failure fields. Deprecated, use Details[\"driver_error\"] to access this.
	DriverError *string `json:"DriverError,omitempty"`
	// DriverMessage indicates a driver action being taken. Deprecated, use Details[\"driver_message\"] to access this.
	DriverMessage *string `json:"DriverMessage,omitempty"`
	// Deprecated, use Details[\"exit_code\"] to access this.
	ExitCode *int64 `json:"ExitCode,omitempty"`
	// Name of the sibling task that caused termination of the task that the TaskEvent refers to. Deprecated, use Details[\"failed_sibling\"] to access this.
	FailedSibling *string `json:"FailedSibling,omitempty"`
	// FailsTask marks whether this event fails the task. Deprecated, use Details[\"fails_task\"] to access this.
	FailsTask *bool `json:"FailsTask,omitempty"`
	// GenericSource is the source of a message. Deprecated, is redundant with event type.
	GenericSource *string `json:"GenericSource,omitempty"`
	// Task Killed Fields. Deprecated, use Details[\"kill_error\"] to access this.
	KillError *string `json:"KillError,omitempty"`
	// KillReason is the reason the task was killed Deprecated, use Details[\"kill_reason\"] to access this.
	KillReason *string `json:"KillReason,omitempty"`
	// A Duration represents the elapsed time between two instants as an int64 nanosecond count. The representation limits the largest representable duration to approximately 290 years.
	KillTimeout *int64 `json:"KillTimeout,omitempty"`
	Message *string `json:"Message,omitempty"`
	// Restart fields. Deprecated, use Details[\"restart_reason\"] to access this.
	RestartReason *string `json:"RestartReason,omitempty"`
	// Setup Failure fields. Deprecated, use Details[\"setup_error\"] to access this.
	SetupError *string `json:"SetupError,omitempty"`
	// Deprecated, use Details[\"signal\"] to access this.
	Signal *int64 `json:"Signal,omitempty"`
	// TaskRestarting fields. Deprecated, use Details[\"start_delay\"] to access this.
	StartDelay *int64 `json:"StartDelay,omitempty"`
	// TaskSignal is the signal that was sent to the task Deprecated, use Details[\"task_signal\"] to access this.
	TaskSignal *string `json:"TaskSignal,omitempty"`
	// TaskSignalReason indicates the reason the task is being signalled. Deprecated, use Details[\"task_signal_reason\"] to access this.
	TaskSignalReason *string `json:"TaskSignalReason,omitempty"`
	Time *int64 `json:"Time,omitempty"`
	Type *string `json:"Type,omitempty"`
	// Validation fields Deprecated, use Details[\"validation_error\"] to access this.
	ValidationError *string `json:"ValidationError,omitempty"`
	// VaultError is the error from token renewal Deprecated, use Details[\"vault_renewal_error\"] to access this.
	VaultError *string `json:"VaultError,omitempty"`
}

// NewTaskEvent instantiates a new TaskEvent object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTaskEvent() *TaskEvent {
	this := TaskEvent{}
	return &this
}

// NewTaskEventWithDefaults instantiates a new TaskEvent object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTaskEventWithDefaults() *TaskEvent {
	this := TaskEvent{}
	return &this
}

// GetDetails returns the Details field value if set, zero value otherwise.
func (o *TaskEvent) GetDetails() map[string]string {
	if o == nil || o.Details == nil {
		var ret map[string]string
		return ret
	}
	return *o.Details
}

// GetDetailsOk returns a tuple with the Details field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetDetailsOk() (*map[string]string, bool) {
	if o == nil || o.Details == nil {
		return nil, false
	}
	return o.Details, true
}

// HasDetails returns a boolean if a field has been set.
func (o *TaskEvent) HasDetails() bool {
	if o != nil && o.Details != nil {
		return true
	}

	return false
}

// SetDetails gets a reference to the given map[string]string and assigns it to the Details field.
func (o *TaskEvent) SetDetails(v map[string]string) {
	o.Details = &v
}

// GetDiskLimit returns the DiskLimit field value if set, zero value otherwise.
func (o *TaskEvent) GetDiskLimit() int64 {
	if o == nil || o.DiskLimit == nil {
		var ret int64
		return ret
	}
	return *o.DiskLimit
}

// GetDiskLimitOk returns a tuple with the DiskLimit field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetDiskLimitOk() (*int64, bool) {
	if o == nil || o.DiskLimit == nil {
		return nil, false
	}
	return o.DiskLimit, true
}

// HasDiskLimit returns a boolean if a field has been set.
func (o *TaskEvent) HasDiskLimit() bool {
	if o != nil && o.DiskLimit != nil {
		return true
	}

	return false
}

// SetDiskLimit gets a reference to the given int64 and assigns it to the DiskLimit field.
func (o *TaskEvent) SetDiskLimit(v int64) {
	o.DiskLimit = &v
}

// GetDisplayMessage returns the DisplayMessage field value if set, zero value otherwise.
func (o *TaskEvent) GetDisplayMessage() string {
	if o == nil || o.DisplayMessage == nil {
		var ret string
		return ret
	}
	return *o.DisplayMessage
}

// GetDisplayMessageOk returns a tuple with the DisplayMessage field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetDisplayMessageOk() (*string, bool) {
	if o == nil || o.DisplayMessage == nil {
		return nil, false
	}
	return o.DisplayMessage, true
}

// HasDisplayMessage returns a boolean if a field has been set.
func (o *TaskEvent) HasDisplayMessage() bool {
	if o != nil && o.DisplayMessage != nil {
		return true
	}

	return false
}

// SetDisplayMessage gets a reference to the given string and assigns it to the DisplayMessage field.
func (o *TaskEvent) SetDisplayMessage(v string) {
	o.DisplayMessage = &v
}

// GetDownloadError returns the DownloadError field value if set, zero value otherwise.
func (o *TaskEvent) GetDownloadError() string {
	if o == nil || o.DownloadError == nil {
		var ret string
		return ret
	}
	return *o.DownloadError
}

// GetDownloadErrorOk returns a tuple with the DownloadError field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetDownloadErrorOk() (*string, bool) {
	if o == nil || o.DownloadError == nil {
		return nil, false
	}
	return o.DownloadError, true
}

// HasDownloadError returns a boolean if a field has been set.
func (o *TaskEvent) HasDownloadError() bool {
	if o != nil && o.DownloadError != nil {
		return true
	}

	return false
}

// SetDownloadError gets a reference to the given string and assigns it to the DownloadError field.
func (o *TaskEvent) SetDownloadError(v string) {
	o.DownloadError = &v
}

// GetDriverError returns the DriverError field value if set, zero value otherwise.
func (o *TaskEvent) GetDriverError() string {
	if o == nil || o.DriverError == nil {
		var ret string
		return ret
	}
	return *o.DriverError
}

// GetDriverErrorOk returns a tuple with the DriverError field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetDriverErrorOk() (*string, bool) {
	if o == nil || o.DriverError == nil {
		return nil, false
	}
	return o.DriverError, true
}

// HasDriverError returns a boolean if a field has been set.
func (o *TaskEvent) HasDriverError() bool {
	if o != nil && o.DriverError != nil {
		return true
	}

	return false
}

// SetDriverError gets a reference to the given string and assigns it to the DriverError field.
func (o *TaskEvent) SetDriverError(v string) {
	o.DriverError = &v
}

// GetDriverMessage returns the DriverMessage field value if set, zero value otherwise.
func (o *TaskEvent) GetDriverMessage() string {
	if o == nil || o.DriverMessage == nil {
		var ret string
		return ret
	}
	return *o.DriverMessage
}

// GetDriverMessageOk returns a tuple with the DriverMessage field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetDriverMessageOk() (*string, bool) {
	if o == nil || o.DriverMessage == nil {
		return nil, false
	}
	return o.DriverMessage, true
}

// HasDriverMessage returns a boolean if a field has been set.
func (o *TaskEvent) HasDriverMessage() bool {
	if o != nil && o.DriverMessage != nil {
		return true
	}

	return false
}

// SetDriverMessage gets a reference to the given string and assigns it to the DriverMessage field.
func (o *TaskEvent) SetDriverMessage(v string) {
	o.DriverMessage = &v
}

// GetExitCode returns the ExitCode field value if set, zero value otherwise.
func (o *TaskEvent) GetExitCode() int64 {
	if o == nil || o.ExitCode == nil {
		var ret int64
		return ret
	}
	return *o.ExitCode
}

// GetExitCodeOk returns a tuple with the ExitCode field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetExitCodeOk() (*int64, bool) {
	if o == nil || o.ExitCode == nil {
		return nil, false
	}
	return o.ExitCode, true
}

// HasExitCode returns a boolean if a field has been set.
func (o *TaskEvent) HasExitCode() bool {
	if o != nil && o.ExitCode != nil {
		return true
	}

	return false
}

// SetExitCode gets a reference to the given int64 and assigns it to the ExitCode field.
func (o *TaskEvent) SetExitCode(v int64) {
	o.ExitCode = &v
}

// GetFailedSibling returns the FailedSibling field value if set, zero value otherwise.
func (o *TaskEvent) GetFailedSibling() string {
	if o == nil || o.FailedSibling == nil {
		var ret string
		return ret
	}
	return *o.FailedSibling
}

// GetFailedSiblingOk returns a tuple with the FailedSibling field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetFailedSiblingOk() (*string, bool) {
	if o == nil || o.FailedSibling == nil {
		return nil, false
	}
	return o.FailedSibling, true
}

// HasFailedSibling returns a boolean if a field has been set.
func (o *TaskEvent) HasFailedSibling() bool {
	if o != nil && o.FailedSibling != nil {
		return true
	}

	return false
}

// SetFailedSibling gets a reference to the given string and assigns it to the FailedSibling field.
func (o *TaskEvent) SetFailedSibling(v string) {
	o.FailedSibling = &v
}

// GetFailsTask returns the FailsTask field value if set, zero value otherwise.
func (o *TaskEvent) GetFailsTask() bool {
	if o == nil || o.FailsTask == nil {
		var ret bool
		return ret
	}
	return *o.FailsTask
}

// GetFailsTaskOk returns a tuple with the FailsTask field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetFailsTaskOk() (*bool, bool) {
	if o == nil || o.FailsTask == nil {
		return nil, false
	}
	return o.FailsTask, true
}

// HasFailsTask returns a boolean if a field has been set.
func (o *TaskEvent) HasFailsTask() bool {
	if o != nil && o.FailsTask != nil {
		return true
	}

	return false
}

// SetFailsTask gets a reference to the given bool and assigns it to the FailsTask field.
func (o *TaskEvent) SetFailsTask(v bool) {
	o.FailsTask = &v
}

// GetGenericSource returns the GenericSource field value if set, zero value otherwise.
func (o *TaskEvent) GetGenericSource() string {
	if o == nil || o.GenericSource == nil {
		var ret string
		return ret
	}
	return *o.GenericSource
}

// GetGenericSourceOk returns a tuple with the GenericSource field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetGenericSourceOk() (*string, bool) {
	if o == nil || o.GenericSource == nil {
		return nil, false
	}
	return o.GenericSource, true
}

// HasGenericSource returns a boolean if a field has been set.
func (o *TaskEvent) HasGenericSource() bool {
	if o != nil && o.GenericSource != nil {
		return true
	}

	return false
}

// SetGenericSource gets a reference to the given string and assigns it to the GenericSource field.
func (o *TaskEvent) SetGenericSource(v string) {
	o.GenericSource = &v
}

// GetKillError returns the KillError field value if set, zero value otherwise.
func (o *TaskEvent) GetKillError() string {
	if o == nil || o.KillError == nil {
		var ret string
		return ret
	}
	return *o.KillError
}

// GetKillErrorOk returns a tuple with the KillError field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetKillErrorOk() (*string, bool) {
	if o == nil || o.KillError == nil {
		return nil, false
	}
	return o.KillError, true
}

// HasKillError returns a boolean if a field has been set.
func (o *TaskEvent) HasKillError() bool {
	if o != nil && o.KillError != nil {
		return true
	}

	return false
}

// SetKillError gets a reference to the given string and assigns it to the KillError field.
func (o *TaskEvent) SetKillError(v string) {
	o.KillError = &v
}

// GetKillReason returns the KillReason field value if set, zero value otherwise.
func (o *TaskEvent) GetKillReason() string {
	if o == nil || o.KillReason == nil {
		var ret string
		return ret
	}
	return *o.KillReason
}

// GetKillReasonOk returns a tuple with the KillReason field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetKillReasonOk() (*string, bool) {
	if o == nil || o.KillReason == nil {
		return nil, false
	}
	return o.KillReason, true
}

// HasKillReason returns a boolean if a field has been set.
func (o *TaskEvent) HasKillReason() bool {
	if o != nil && o.KillReason != nil {
		return true
	}

	return false
}

// SetKillReason gets a reference to the given string and assigns it to the KillReason field.
func (o *TaskEvent) SetKillReason(v string) {
	o.KillReason = &v
}

// GetKillTimeout returns the KillTimeout field value if set, zero value otherwise.
func (o *TaskEvent) GetKillTimeout() int64 {
	if o == nil || o.KillTimeout == nil {
		var ret int64
		return ret
	}
	return *o.KillTimeout
}

// GetKillTimeoutOk returns a tuple with the KillTimeout field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetKillTimeoutOk() (*int64, bool) {
	if o == nil || o.KillTimeout == nil {
		return nil, false
	}
	return o.KillTimeout, true
}

// HasKillTimeout returns a boolean if a field has been set.
func (o *TaskEvent) HasKillTimeout() bool {
	if o != nil && o.KillTimeout != nil {
		return true
	}

	return false
}

// SetKillTimeout gets a reference to the given int64 and assigns it to the KillTimeout field.
func (o *TaskEvent) SetKillTimeout(v int64) {
	o.KillTimeout = &v
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *TaskEvent) GetMessage() string {
	if o == nil || o.Message == nil {
		var ret string
		return ret
	}
	return *o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetMessageOk() (*string, bool) {
	if o == nil || o.Message == nil {
		return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *TaskEvent) HasMessage() bool {
	if o != nil && o.Message != nil {
		return true
	}

	return false
}

// SetMessage gets a reference to the given string and assigns it to the Message field.
func (o *TaskEvent) SetMessage(v string) {
	o.Message = &v
}

// GetRestartReason returns the RestartReason field value if set, zero value otherwise.
func (o *TaskEvent) GetRestartReason() string {
	if o == nil || o.RestartReason == nil {
		var ret string
		return ret
	}
	return *o.RestartReason
}

// GetRestartReasonOk returns a tuple with the RestartReason field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetRestartReasonOk() (*string, bool) {
	if o == nil || o.RestartReason == nil {
		return nil, false
	}
	return o.RestartReason, true
}

// HasRestartReason returns a boolean if a field has been set.
func (o *TaskEvent) HasRestartReason() bool {
	if o != nil && o.RestartReason != nil {
		return true
	}

	return false
}

// SetRestartReason gets a reference to the given string and assigns it to the RestartReason field.
func (o *TaskEvent) SetRestartReason(v string) {
	o.RestartReason = &v
}

// GetSetupError returns the SetupError field value if set, zero value otherwise.
func (o *TaskEvent) GetSetupError() string {
	if o == nil || o.SetupError == nil {
		var ret string
		return ret
	}
	return *o.SetupError
}

// GetSetupErrorOk returns a tuple with the SetupError field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetSetupErrorOk() (*string, bool) {
	if o == nil || o.SetupError == nil {
		return nil, false
	}
	return o.SetupError, true
}

// HasSetupError returns a boolean if a field has been set.
func (o *TaskEvent) HasSetupError() bool {
	if o != nil && o.SetupError != nil {
		return true
	}

	return false
}

// SetSetupError gets a reference to the given string and assigns it to the SetupError field.
func (o *TaskEvent) SetSetupError(v string) {
	o.SetupError = &v
}

// GetSignal returns the Signal field value if set, zero value otherwise.
func (o *TaskEvent) GetSignal() int64 {
	if o == nil || o.Signal == nil {
		var ret int64
		return ret
	}
	return *o.Signal
}

// GetSignalOk returns a tuple with the Signal field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetSignalOk() (*int64, bool) {
	if o == nil || o.Signal == nil {
		return nil, false
	}
	return o.Signal, true
}

// HasSignal returns a boolean if a field has been set.
func (o *TaskEvent) HasSignal() bool {
	if o != nil && o.Signal != nil {
		return true
	}

	return false
}

// SetSignal gets a reference to the given int64 and assigns it to the Signal field.
func (o *TaskEvent) SetSignal(v int64) {
	o.Signal = &v
}

// GetStartDelay returns the StartDelay field value if set, zero value otherwise.
func (o *TaskEvent) GetStartDelay() int64 {
	if o == nil || o.StartDelay == nil {
		var ret int64
		return ret
	}
	return *o.StartDelay
}

// GetStartDelayOk returns a tuple with the StartDelay field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetStartDelayOk() (*int64, bool) {
	if o == nil || o.StartDelay == nil {
		return nil, false
	}
	return o.StartDelay, true
}

// HasStartDelay returns a boolean if a field has been set.
func (o *TaskEvent) HasStartDelay() bool {
	if o != nil && o.StartDelay != nil {
		return true
	}

	return false
}

// SetStartDelay gets a reference to the given int64 and assigns it to the StartDelay field.
func (o *TaskEvent) SetStartDelay(v int64) {
	o.StartDelay = &v
}

// GetTaskSignal returns the TaskSignal field value if set, zero value otherwise.
func (o *TaskEvent) GetTaskSignal() string {
	if o == nil || o.TaskSignal == nil {
		var ret string
		return ret
	}
	return *o.TaskSignal
}

// GetTaskSignalOk returns a tuple with the TaskSignal field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetTaskSignalOk() (*string, bool) {
	if o == nil || o.TaskSignal == nil {
		return nil, false
	}
	return o.TaskSignal, true
}

// HasTaskSignal returns a boolean if a field has been set.
func (o *TaskEvent) HasTaskSignal() bool {
	if o != nil && o.TaskSignal != nil {
		return true
	}

	return false
}

// SetTaskSignal gets a reference to the given string and assigns it to the TaskSignal field.
func (o *TaskEvent) SetTaskSignal(v string) {
	o.TaskSignal = &v
}

// GetTaskSignalReason returns the TaskSignalReason field value if set, zero value otherwise.
func (o *TaskEvent) GetTaskSignalReason() string {
	if o == nil || o.TaskSignalReason == nil {
		var ret string
		return ret
	}
	return *o.TaskSignalReason
}

// GetTaskSignalReasonOk returns a tuple with the TaskSignalReason field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetTaskSignalReasonOk() (*string, bool) {
	if o == nil || o.TaskSignalReason == nil {
		return nil, false
	}
	return o.TaskSignalReason, true
}

// HasTaskSignalReason returns a boolean if a field has been set.
func (o *TaskEvent) HasTaskSignalReason() bool {
	if o != nil && o.TaskSignalReason != nil {
		return true
	}

	return false
}

// SetTaskSignalReason gets a reference to the given string and assigns it to the TaskSignalReason field.
func (o *TaskEvent) SetTaskSignalReason(v string) {
	o.TaskSignalReason = &v
}

// GetTime returns the Time field value if set, zero value otherwise.
func (o *TaskEvent) GetTime() int64 {
	if o == nil || o.Time == nil {
		var ret int64
		return ret
	}
	return *o.Time
}

// GetTimeOk returns a tuple with the Time field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetTimeOk() (*int64, bool) {
	if o == nil || o.Time == nil {
		return nil, false
	}
	return o.Time, true
}

// HasTime returns a boolean if a field has been set.
func (o *TaskEvent) HasTime() bool {
	if o != nil && o.Time != nil {
		return true
	}

	return false
}

// SetTime gets a reference to the given int64 and assigns it to the Time field.
func (o *TaskEvent) SetTime(v int64) {
	o.Time = &v
}

// GetType returns the Type field value if set, zero value otherwise.
func (o *TaskEvent) GetType() string {
	if o == nil || o.Type == nil {
		var ret string
		return ret
	}
	return *o.Type
}

// GetTypeOk returns a tuple with the Type field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetTypeOk() (*string, bool) {
	if o == nil || o.Type == nil {
		return nil, false
	}
	return o.Type, true
}

// HasType returns a boolean if a field has been set.
func (o *TaskEvent) HasType() bool {
	if o != nil && o.Type != nil {
		return true
	}

	return false
}

// SetType gets a reference to the given string and assigns it to the Type field.
func (o *TaskEvent) SetType(v string) {
	o.Type = &v
}

// GetValidationError returns the ValidationError field value if set, zero value otherwise.
func (o *TaskEvent) GetValidationError() string {
	if o == nil || o.ValidationError == nil {
		var ret string
		return ret
	}
	return *o.ValidationError
}

// GetValidationErrorOk returns a tuple with the ValidationError field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetValidationErrorOk() (*string, bool) {
	if o == nil || o.ValidationError == nil {
		return nil, false
	}
	return o.ValidationError, true
}

// HasValidationError returns a boolean if a field has been set.
func (o *TaskEvent) HasValidationError() bool {
	if o != nil && o.ValidationError != nil {
		return true
	}

	return false
}

// SetValidationError gets a reference to the given string and assigns it to the ValidationError field.
func (o *TaskEvent) SetValidationError(v string) {
	o.ValidationError = &v
}

// GetVaultError returns the VaultError field value if set, zero value otherwise.
func (o *TaskEvent) GetVaultError() string {
	if o == nil || o.VaultError == nil {
		var ret string
		return ret
	}
	return *o.VaultError
}

// GetVaultErrorOk returns a tuple with the VaultError field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TaskEvent) GetVaultErrorOk() (*string, bool) {
	if o == nil || o.VaultError == nil {
		return nil, false
	}
	return o.VaultError, true
}

// HasVaultError returns a boolean if a field has been set.
func (o *TaskEvent) HasVaultError() bool {
	if o != nil && o.VaultError != nil {
		return true
	}

	return false
}

// SetVaultError gets a reference to the given string and assigns it to the VaultError field.
func (o *TaskEvent) SetVaultError(v string) {
	o.VaultError = &v
}

func (o TaskEvent) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Details != nil {
		toSerialize["Details"] = o.Details
	}
	if o.DiskLimit != nil {
		toSerialize["DiskLimit"] = o.DiskLimit
	}
	if o.DisplayMessage != nil {
		toSerialize["DisplayMessage"] = o.DisplayMessage
	}
	if o.DownloadError != nil {
		toSerialize["DownloadError"] = o.DownloadError
	}
	if o.DriverError != nil {
		toSerialize["DriverError"] = o.DriverError
	}
	if o.DriverMessage != nil {
		toSerialize["DriverMessage"] = o.DriverMessage
	}
	if o.ExitCode != nil {
		toSerialize["ExitCode"] = o.ExitCode
	}
	if o.FailedSibling != nil {
		toSerialize["FailedSibling"] = o.FailedSibling
	}
	if o.FailsTask != nil {
		toSerialize["FailsTask"] = o.FailsTask
	}
	if o.GenericSource != nil {
		toSerialize["GenericSource"] = o.GenericSource
	}
	if o.KillError != nil {
		toSerialize["KillError"] = o.KillError
	}
	if o.KillReason != nil {
		toSerialize["KillReason"] = o.KillReason
	}
	if o.KillTimeout != nil {
		toSerialize["KillTimeout"] = o.KillTimeout
	}
	if o.Message != nil {
		toSerialize["Message"] = o.Message
	}
	if o.RestartReason != nil {
		toSerialize["RestartReason"] = o.RestartReason
	}
	if o.SetupError != nil {
		toSerialize["SetupError"] = o.SetupError
	}
	if o.Signal != nil {
		toSerialize["Signal"] = o.Signal
	}
	if o.StartDelay != nil {
		toSerialize["StartDelay"] = o.StartDelay
	}
	if o.TaskSignal != nil {
		toSerialize["TaskSignal"] = o.TaskSignal
	}
	if o.TaskSignalReason != nil {
		toSerialize["TaskSignalReason"] = o.TaskSignalReason
	}
	if o.Time != nil {
		toSerialize["Time"] = o.Time
	}
	if o.Type != nil {
		toSerialize["Type"] = o.Type
	}
	if o.ValidationError != nil {
		toSerialize["ValidationError"] = o.ValidationError
	}
	if o.VaultError != nil {
		toSerialize["VaultError"] = o.VaultError
	}
	return json.Marshal(toSerialize)
}

type NullableTaskEvent struct {
	value *TaskEvent
	isSet bool
}

func (v NullableTaskEvent) Get() *TaskEvent {
	return v.value
}

func (v *NullableTaskEvent) Set(val *TaskEvent) {
	v.value = val
	v.isSet = true
}

func (v NullableTaskEvent) IsSet() bool {
	return v.isSet
}

func (v *NullableTaskEvent) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTaskEvent(val *TaskEvent) *NullableTaskEvent {
	return &NullableTaskEvent{value: val, isSet: true}
}

func (v NullableTaskEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTaskEvent) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



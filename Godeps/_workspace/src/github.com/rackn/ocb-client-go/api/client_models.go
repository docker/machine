package crowbar
// Apache 2 License 2015 by Rob Hirschfeld for RackN

type CrowbarDigest struct {
    CSRFToken string `json:"csrf_token"`
    Message   string `json:"message"`
}

type Deployment struct {
    ID          int64   `json:"id"`
    State       int     `json:"state"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    System      bool    `json:"system"`
    ParentID    int64   `json:"parent_id"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

type NewDeployment struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    ParentID    int64   `json:"parent_id"`
}


type Node struct {
    ID           int64         `json:"id"`
    Name         string        `json:"name"`
    Description  string        `json:"description"`
    Admin        bool          `json:"admin"`
    Alias        string        `json:"alias"`
    Alive        bool          `json:"alive"`
    Allocated    bool          `json:"allocated"`
    Available    bool          `json:"available"`
    Bootenv      string        `json:"bootenv"`
    DeploymentID int64         `json:"deployment_id"`
    Order        int64         `json:"order"`
    System       bool          `json:"system"`
    TargetRoleID int64         `json:"target_role_id"`
    Discovery    interface{}   `json:"discovery"`
    CreatedAt    string        `json:"created_at"`
    UpdatedAt    string        `json:"updated_at"`
}

type NewNode struct {
    Name         string        `json:"name"`
    Description  string        `json:"description"`
    Alias        string        `json:"alias"`
    Available    bool          `json:"available"`
    Bootenv      string        `json:"bootenv"`
    DeploymentID int64         `json:"deployment_id"`
    Order        int64         `json:"order"`
}


type NodePower struct {
    ID              int64       `json:"id"`
    Action          string      `json:"action"`
    Result          string      `json:"result"`
}

type NodeRole struct {
    ID              int64       `json:"id"`
    DeploymentID    int64       `json:"deployment_id"`
    RoleID          int64       `json:"role_id"`
    NodeID          int64       `json:"node_id"`
    State           int         `json:"state"`
    Status          string      `json:"status"`
    RunLog          string      `json:"runlog"`
    Available       bool        `json:"available"`
    Order           int         `json:"order"`
    ProposedData    interface{}    `json:"proposed_data"`
    CommittedData   interface{}    `json:"committed_data"`
    SysData         interface{}    `json:"sysdata"`
    Wall            interface{}    `json:"wall"`
    NodeError       bool        `json:"node_error"`
    CreatedAt       string      `json:"created_at"`
    UpdatedAt       string      `json:"updated_at"`
}

// designed to support Crowbar_Access, but could be used for others
type NodeAttrib struct {
    ID              int64       `json:"id"`
    Name            string      `json:"name"`
    Description     string      `json:"description"`
    BarclampID      int64       `json:"barclamp_id"`
    RoleID          int64       `json:"role_id"`
    NodeID          int64       `json:"node_id"`
    Writable        bool        `json:"writable"`
    Schema          interface{} `json:"schema"`
    Order           int         `json:"order"`
    Map             string      `json:"map"`
    UIRenderer      string      `json:"ui_renderer"`
    Value           map[string]string      `json:"value"`
    CreatedAt       string      `json:"created_at"`
    UpdatedAt       string      `json:"updated_at"`
}

// designed to support Crowbar_Access, but could be used for others
type NodeAttribValue struct {
    Value           map[string]string      `json:"value"`
}

type NodeAddress struct {
    Node      string    `json:"node"`
    Network   string    `json:"network"`
    Category  string    `json:"category"`
    Addresses []string  `json:"addresses"`
}

type NewNodeRole struct {
    DeploymentID    int64       `json:"deployment_id"`
    RoleID          int64       `json:"role_id"`
    NodeID          int64       `json:"node_id"`
    Order           int         `json:"order"`
    ProposedData    interface{} `json:"proposed_data"`
}

type Role struct {
    ID          int64         `json:"id"`
    Name        string        `json:"name"`
    Description string        `json:"description"`
    BarclampID  int64         `json:"barclamp_id"`
    JigName     string        `json:"jig_name"`
    Abstract    bool          `json:"abstract"`
    Bootstrap   bool          `json:"bootstrap"`
    Cluster     bool          `json:"cluster"`
    Destructive bool          `json:"destructive"`
    Discovery   bool          `json:"discovery"`
    Implicit    bool          `json:"implicit"`
    Library     bool          `json:"library"`
    Milestone   bool          `json:"milestone"`
    Powersave   bool          `json:"powersave"`
    Service     bool          `json:"service"`
    Cohort      int           `json:"cohort"`
    Conflicts   []interface{} `json:"conflicts"`
    Provides    []interface{} `json:"provides"`
    Template    struct{}      `json:"template"`
    CreatedAt   string        `json:"created_at"`
    UpdatedAt   string        `json:"updated_at"`
}


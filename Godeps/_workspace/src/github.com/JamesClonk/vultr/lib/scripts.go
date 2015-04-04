package lib

import "net/url"

// StartupScript on Vultr account
type StartupScript struct {
	ID      string `json:"SCRIPTID"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"script"`
}

func (c *Client) GetStartupScripts() (scripts []StartupScript, err error) {
	var scriptMap map[string]StartupScript
	if err := c.get(`startupscript/list`, &scriptMap); err != nil {
		return nil, err
	}

	for _, script := range scriptMap {
		if script.Type == "" {
			script.Type = "boot" // set default script type
		}
		scripts = append(scripts, script)
	}
	return scripts, nil
}

func (c *Client) GetStartupScript(id string) (StartupScript, error) {
	scripts, err := c.GetStartupScripts()
	if err != nil {
		return StartupScript{}, err
	}

	for _, s := range scripts {
		if s.ID == id {
			return s, nil
		}
	}
	return StartupScript{}, nil
}

func (c *Client) CreateStartupScript(name, content, scriptType string) (StartupScript, error) {
	values := url.Values{
		"name":   {name},
		"script": {content},
		"type":   {scriptType},
	}

	var script StartupScript
	if err := c.post(`startupscript/create`, values, &script); err != nil {
		return StartupScript{}, err
	}
	script.Name = name
	script.Content = content
	script.Type = scriptType

	return script, nil
}

func (c *Client) UpdateStartupScript(script StartupScript) error {
	values := url.Values{
		"SCRIPTID": {script.ID},
	}
	if script.Name != "" {
		values.Add("name", script.Name)
	}
	if script.Content != "" {
		values.Add("script", script.Content)
	}

	if err := c.post(`startupscript/update`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteStartupScript(id string) error {
	values := url.Values{
		"SCRIPTID": {id},
	}

	if err := c.post(`startupscript/destroy`, values, nil); err != nil {
		return err
	}
	return nil
}

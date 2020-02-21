struct ConfigParamValue {
    1: string key
    2: string name
    3: string type
    4: list<string> valuesList
    5: string value
}

service AppConfigService {
    list<ConfigParamValue> getParamValues()
    void setParamValue(1:string key, 2:string value)
    string getParamValue(1:string key)
}

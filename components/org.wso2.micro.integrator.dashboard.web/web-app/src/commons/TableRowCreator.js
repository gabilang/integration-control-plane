/*
 * Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 *
 */

import React from 'react';
import axios from 'axios';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import Switch from "react-switch";
import { makeStyles } from '@material-ui/core/styles';
import { useSelector } from 'react-redux';
import EnabledIcon from '@material-ui/icons/CheckCircleOutlineOutlined';
import DisabledIcon from '@material-ui/icons/BlockOutlined';
import InputLabel from '@material-ui/core/InputLabel';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import { Button } from '@material-ui/core';
import NodesCell from './NodesCell';
import LogsNodeCell from './LogsNodeCell';
import StatusCell from '../commons/StatusCell';
import UserIdCell from '../commons/UserIdCell';

export default function TableRowCreator(props) {
    const { pageId, data, headers } = props;
    return <TableRow>
        {headers.map(header => {switch(header.id) {
            // common
            case 'name':
                return <TableCell>{data.name}</TableCell>
            case 'nodes':
                return <TableCell><table>{data.nodes.map(node=><NodesCell pageId={pageId} nodeData={node} />)}</table></TableCell>
            case 'type':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.type} />)}</table></TableCell>

            // Proxy Services and dataservices
            case 'wsdlUrl':
                return <TableCell><table>{data.nodes.map(node=><LinkCell data={node.details.wsdl1_1} />)}</table></TableCell>

            // Proxy Services
            case 'isRunning':
                return <TableCell>{data.nodes.map(node=><SwitchStatusCell pageId={pageId} artifactName={node.details.name} 
                        nodeId={node.nodeId} status={node.details.isRunning}/>)}</TableCell>
            // Endpoints
            case 'state':
                return <TableCell>{data.nodes.map(node=><SwitchStatusCell pageId={pageId} artifactName={node.details.name} 
                        nodeId={node.nodeId} status={node.details.isActive}/>)}</TableCell>

            // Inbound Endpoints
            case 'protocol':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.protocol} />)}</table></TableCell>

            // Message Stores
            case 'message_count':
                return <TableCell>{data.nodes.map(node=><StringCell data={node.details.size === '-1' ? "Not Supported" : node.details.size}/>)}</TableCell>

            // Message Processors
            case 'status':
                return <TableCell>{data.nodes.map(node=><SwitchStatusCell pageId={pageId} artifactName={node.details.name} 
                        nodeId={node.nodeId} status={node.details.status === 'active' ? true : false}/>)}</TableCell>

            // Apis
            case 'url':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.url} />)}</table></TableCell>

            // Templates and Connectors
            case 'template_nodes':
            case 'connector_nodes':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.nodeId} />)}</table></TableCell>

            // Sequences
            case 'statistic':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.stats} />)}</table></TableCell>
            
            // Tasks
            case 'group':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.taskGroup} />)}</table></TableCell>

            // Connectors
            case 'package':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.package} />)}</table></TableCell>
            case 'description':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.description} />)}</table></TableCell>
            case 'connector_status':
                return <TableCell><table>{data.nodes.map(node=><ConnectorStatus status={node.details.status} />)}</table></TableCell>

            //carbon apps
            case 'version':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.details.version} />)}</table></TableCell>

            case 'size':
                return <TableCell><table>{data.nodes.map(node=><StringCell data={node.size} />)}</table></TableCell>
            
            case 'data_source_status':
                return <TableCell><table>{data.nodes.map(node=><StatusCell data={node.data_source_status} />)}</table></TableCell>
            case 'port':
                return <TableCell>{data.nodes.map(node=><StringCell data={node.port}/>)}</TableCell>

            // log-configs page
            case 'componentName':
                return <TableCell><StringCell data={data.componentName}/></TableCell>
            
            case 'level':
                return <TableCell><LogConfigLevelDrowDown name={data.name} level={data.level}/></TableCell>

            // users page
            case 'userId':
                return <TableCell><table><UserIdCell id={data} /></table></TableCell>
            
            // Node page
            case 'nodeId':
                return <TableCell><table><NodesCell pageId={pageId} nodeData={data}/></table></TableCell>
            case 'node_status':
                return <TableCell>Active</TableCell>
            case 'role':
                return <TableCell>Member</TableCell>
            case 'nodes_logs':
                return <TableCell><table>{data.nodes.map(node=><LogsNodeCell nodeId={node} fileName={data.name} />)}</table></TableCell>
            default:
                <TableCell>Table data not available</TableCell>
        }})}
    </TableRow>
}

function StringCell(props) {
    var data = props.data
    return <tr><td>{data}</td></tr>
}

function LinkCell(props) {
    const classes = useStyles();
    var data = props.data
    return <tr><td><a className={classes.tableCell} href={data}>{data}</a></td></tr>
}

function ConnectorStatus(props) {
    return (
        <tr>
            {props.status === 'enabled' ? <EnabledIcon style={{color:"green"}}/> : <DisabledIcon style={{color:"red"}}/>}
        </tr>
    )
}

function SwitchStatusCell(props) {
    const { pageId, artifactName, nodeId, status } = props;
    var isActive = status;
    const globalGroupId = useSelector(state => state.groupId);
    const basePath = useSelector(state => state.basePath);

    const changeState = () => {
        isActive = !isActive
        updateArtifact()
    };

    const updateArtifact = () => {
        const url = basePath.concat('/groups/').concat(globalGroupId).concat("/").concat(pageId);
        axios.patch(url, {
            "artifactName": artifactName,
            "nodeId": nodeId,
            "type": "status",
            "value": isActive
        });
    }

    return <tr><td><Switch checked={isActive} onChange={changeState} height={16} width={36} /></td></tr>
}

function LogConfigLevelDrowDown(props) {
    const classes = useStyles();
    const name = props.name;
    const globalGroupId = useSelector(state => state.groupId);
    const basePath = useSelector(state => state.basePath);
    const [ level, setLevel ] = React.useState(props.level);
    const [dialog, setDialog] = React.useState({
        open : false,
        title: '',
        message: ''
    });
    const handleDialogClose = () => {
        setDialog({
            open: false,
            title: '',
            message: ''
        })
    }
    const changeLogLevel = (loggerLevel) => {
        const url = basePath.concat('/groups/').concat(globalGroupId).concat("/log-configs");
        axios.patch(url, {
            "name": name,
            "level": loggerLevel
        }).then(response => {
            if (response.data.status === 'fail') {
                setDialog({
                    open: true, 
                    title: 'Error',
                    message: response.data.message
                })
            }
        })
        setLevel(loggerLevel);
    };
    
    return <FormControl variant="outlined" className={classes.formControl}>
                <InputLabel>Level</InputLabel>
                    <Select
                        native
                        value={level}
                        onChange={(e) => changeLogLevel(e.target.value)}
                        label="Level"
                    >
                        <option value={'OFF'}>OFF</option>
                        <option value={'TRACE'}>TRACE</option>
                        <option value={'DEBUG'}>DEBUG</option>
                        <option value={'INFO'}>INFO</option>
                        <option value={'WARN'}>WARN</option>
                        <option value={'ERROR'}>ERROR</option>
                        <option value={'FATAL'}>FATAL</option>
                    </Select>
                    <Dialog open={dialog.open} onClose={() => handleDialogClose()}
                                    aria-labelledby="alert-dialog-title" aria-describedby="alert-dialog-description">
                        <DialogTitle id="alert-dialog-title">{dialog.title}</DialogTitle>
                        <DialogContent dividers>
                            <DialogContentText id="alert-dialog-description">
                                {dialog.message}
                            </DialogContentText>
                        </DialogContent>
                        <DialogActions>
                            <Button onClick={() => handleDialogClose()} variant="contained" autoFocus>
                                OK
                            </Button>
                        </DialogActions>
                    </Dialog>
            </FormControl>
}

const useStyles = makeStyles((theme) => ({
    tableCell : {
        paddingLeft: '15px',
        color: '#3f51b5'
    },
    formControl: {
        minWidth: 10,
        minHeight: 10
    }
}));

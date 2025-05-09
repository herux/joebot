"use strict";

let clients = [];

new Vue({
	el: '#clients_list',

	data: {
		isLoggedIn: false,
		isBulkIntallView: true,
		isUsersView: false,
		items: clients,
		userArray: [],
		fields: [
			{ key: 'id', label: 'ID', sortable: true, tdAttr: {style:"width:9%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'ip', label: 'IP', sortable: true, tdAttr: {style:"width:8%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'host_name', label: 'Hostname', sortable: true, tdAttr: {style:"width:10%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'username', label: 'User', sortable: true, tdAttr: {style:"width:5%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'port_tunnels', label: 'Port Tunnels', sortable: false, tdAttr: {style:"width:15%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'tags', label: 'Tags', sortable: false, tdAttr: {style:"width:30%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'actions', label: 'Actions', tdAttr: {style:"width:23%;word-break:break-all;word-wrap:break-word;"} }
		],
		userFields: [
			{ key: 'ID', label: 'ID', sortable: true, tdAttr: {style:"width:9%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'Username', label: 'Username', sortable: true, tdAttr: {style:"width:9%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'IsAdmin', label: 'Admin', sortable: false, tdAttr: {style:"width:9%;word-break:break-all;word-wrap:break-word;"} },
			{ key: 'IpWhitelisted', label: 'IP Whitelisted', sortable: false, tdAttr: {style:"width:9%;word-break:break-all;word-wrap:break-word;"} },
		],
		currentPage: 1,
		perPage: 99999,
		totalRows: clients.length,
		sortBy: null,
		sortDesc: false,
		filter: null,
		modalInfo: { title: '', content: '' },
		
		filteredItems: [],
		selected: [],
		selectAll: false,

		targetClientId: null,
		targetClientPortToBeCreated: null,

		targetJoebotServerAddr: null,
		targetIPList: null,
		targetSSHUser: null,
		targetSSHPassword: null,
		targetSSHKeyContent: null,

		email: null,
		password: null,
		loginError: '',
		isadmin: false,
	},

	mounted: function() {
		this.$nextTick(() => {
			if (!this.isLoggedIn) {
				this.login();
			}
		});
	},

	watch: function() {
		console.log('watch')
	},
	
	created: function() {
		const token = this.getCookie("authToken");
		if (token) {
			this.isLoggedIn = true; 
		} else {
			this.$nextTick(() => this.$refs.modalLogin.show());
		}

		const updateUserTable = () => {
			const token = this.getCookie("authToken");
			this.$http.get('/api/users', {
				headers: {
					Authorization: `Bearer ${token}`
				}
			}).then(result => {
				this.userArray = result.body;
			});
		};

		const updateTable = () => {
			this.$http.get('/api/clients', {
				headers: {
					Authorization: `Bearer ${this.getCookie("authToken")}`
				}
			}).then(result => {
				let clients = result.body.clients;
				let newClientIDs = clients.map(client => client.id);
				
				//Remove existing entries
				this.items = this.items.filter(item => newClientIDs.indexOf(item.id) >= 0);
				
				//Update existing entries
				this.items.forEach((item, index) => {
					let obj = clients.find(client => client.id == item.id);
					Object.keys(obj).forEach((key, index) => {
						item[key] = obj[key];
					});
				});
				
				// Append new clients
				clients.forEach(client => {
					let itemIds = this.items.map(item => item.id);
					if( itemIds.indexOf(client.id) < 0 ){
						this.items.push(client);
					}
				});
				
				this.totalRows = this.items.length;
			});
		};
		
		let url = window.location.href;
		let captured = /filter=([^&]+)/.exec(url);
		if (captured){
			this.filter = captured[1] ? decodeURIComponent(captured[1]) : null;
		}

		// updateTable();
		setInterval(() => { 
			if (this.isLoggedIn){
				updateTable();
				updateUserTable();
			}
		}, 1000);
	},
	
	methods: {
		select () {
			let visibleItemIDs = this.filteredItems.map(item => item.id);
			this.selected = [];
			if (!this.selectAll) {
				for (let i in this.items) {
					if (this.filteredItems.length === 0 || visibleItemIDs.includes(this.items[i].id))
						this.selected.push(this.items[i].id);
				}
			}
		},
		info (item, index, button) {
			this.modalInfo.title = item.id;
			this.modalInfo.content = JSON.stringify(item, null, 3);
			this.$root.$emit('bv::show::modal', 'modalInfo', button);
		},
		open_terminal (item) {
			if( item.gotty_web_terminal_info ){
				window.open(`http://${window.location.hostname}:${item.gotty_web_terminal_info.port_tunnel.server_port}`);
			}
		},
		open_filebrowser (item) {
			if( item.filebrowser_info ){
				var postURL = "files" + encodeURI(item.filebrowser_info.default_directory);
				window.open(`http://${window.location.hostname}:${item.filebrowser_info.port_tunnel.server_port}/${postURL}`);
			}
		},
		open_vnc (item) {
			if( item.novnc_websocket_info ){
				window.open(`http://novnc.com/noVNC/vnc.html?host=${window.location.hostname}&port=${item.novnc_websocket_info.port_tunnel.server_port}&encrypt=0&autoconnect=1`);
			}
		},
		resetModal () {
			this.modalInfo.title = '';
			this.modalInfo.content = '';
		},
		onFiltered (filteredItems) {
			this.filteredItems = filteredItems;
			this.totalRows = filteredItems.length;
			this.currentPage = 1;
		},
		users () {
			this.isBulkIntallView = false;
			this.isUsersView = true;
			this.$refs.modalUser.show();
		},
		addUser () {
			this.$refs.modalUser.show();
		},
		handleAddUserOk () {
			let data = new FormData();
			data.set("username", this.email);
			data.set("password", this.password);

			this.$http.post(`/api/users`, data, {
				headers: {
					Authorization: `Bearer ${this.getCookie("authToken")}`
				}
			}).then((response) => {
				if (response.body && response.body.token) {
					this.$refs.modalUser.hide();
				} else {
					this.loginError = "Add user failed";
				}
			});
		},
		bulk_install () {
			this.isBulkIntallView = true;
			this.isUsersView = false;
			this.targetJoebotServerAddr = location.hostname + ":13579",
			this.targetIPList = "";
			this.targetSSHUser = "";
			this.targetSSHPassword = "";
			this.targetSSHKeyContent = "";
			this.$refs.modalInitBulkInstall.show();
			console.log('bulk_install: ', this.isBulkIntallView, this.isUsersView);
		},
		handleInitBulkInstallOk (evt) {
			let bulkInstallInfo = {
				'JoebotServerIP': null,
				'JoebotServerPort': 13579,
				'Addresses': [],
				'Username': this.targetSSHUser,
				'Password': this.targetSSHPassword,
				'Key': this.targetSSHKeyContent
			};

			if (this.targetJoebotServerAddr === ''){
				alert('Missing joebot server address');
				return
			}
			if (this.targetSSHUser === ''){
				alert('SSH username is missing');
				return;
			}
			if (this.targetSSHPassword === '' && this.targetSSHKeyContent === ''){
				alert('Either SSH password or key must be defined');
				return;
			}

			let joebotAddr = this.targetJoebotServerAddr.split(':');
			bulkInstallInfo.JoebotServerIP = joebotAddr[0];
			bulkInstallInfo.JoebotServerPort = (joebotAddr.length===2)?parseInt(joebotAddr[1]):bulkInstallInfo.JoebotServerPort;

			let addresses = this.targetIPList.split('\n');
			for (let address of addresses) {
				address = address.trim();
				if (address === '')
					continue;

				let tmp = address.trim().split(':');
				let ip = tmp[0]
				let port = (tmp.length === 2)?parseInt(tmp[1]):22;

				bulkInstallInfo['Addresses'].push({
					'IP': ip,
					'Port': port
				});
			}

			if (bulkInstallInfo['Addresses'].length === 0){
				alert('Address list is empty');
				return;
			}

			this.handleSubmitInitBulkInstall(bulkInstallInfo);
		},
		handleSubmitInitBulkInstall (bulkInstallInfo) {
			axios.post('/api/bulk-install', bulkInstallInfo, {
				headers: {
					Authorization: `Bearer ${this.getCookie("authToken")}`
				}
			})
			.then(function (response) {
				console.log(response);
			})
			.catch(function (error) {
				console.log(error);
			});

			this.$refs.modalInitBulkInstall.hide();
		},
		focusTargetIPList () {
			this.$refs.modalTargetIPList.focus();
		},

		create_tunnel (item) {
			this.targetClientId = item.id;
			this.$refs.modalCreateTunnel.show();
		},
		focusInputPort () {
			this.$refs.modalTargetPortInput.focus();
		},
		handleCreateTunnelOk (evt) {
			if (!this.targetClientPortToBeCreated || parseInt(this.targetClientPortToBeCreated) <= 0) {
				evt.preventDefault();
				this.targetClientPortToBeCreated = null;
				alert('Please enter a valid client port');
			} else {
				this.handleSubmitTunnelCreation();
			}
		},
		handleSubmitTunnelCreation () {
			let data = new FormData();
			data.set('target_client_port', parseInt(this.targetClientPortToBeCreated))
			this.$http.post(`/api/client/${this.targetClientId}`, data, {
				headers: {
					Authorization: `Bearer ${this.getCookie("authToken")}`
				}
			}).then(response => {
				console.log(`Created Tunnel: \n${JSON.stringify(response.body, null, 3)}`);
			});
			
			this.$refs.modalCreateTunnel.hide();
			this.targetClientId = null;
			this.targetClientPortToBeCreated = null;
		},
		login () {
			if (this.$refs.modalLogin) {
				this.email = "";
				this.password = "";
				this.$refs.modalLogin.show();
			}
		},
		handleLoginOk (evt) {
			let data = new FormData();
			data.set("username", this.email);
			data.set("password", this.password);

			this.$http.post(`/api/login`, data).then((response) => {
				if (response.status === 200 && response.body && response.body.token) {
					this.setCookie("authToken", response.body.token, 7); 
					this.isLoggedIn = true;
					this.$refs.modalLogin.hide();
				} else {
					this.loginError = "Login failed. Please check your credentials.";
				}
			}).catch((error) => {
				this.loginError = "Login failed. Please check your credentials.";
				console.error(error);
			});
		},
		logout() {
			this.deleteCookie("authToken");
			this.isLoggedIn = false;
			this.$nextTick(() => this.$refs.modalLogin.show());
		},
		setCookie(name, value, days) {
			const date = new Date();
			date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
			document.cookie = `${name}=${value};expires=${date.toUTCString()};path=/`;
		},
		getCookie(name) {
			const value = `; ${document.cookie}`;
			const parts = value.split(`; ${name}=`);
			if (parts.length === 2) return parts.pop().split(";").shift();
		},
		deleteCookie(name) {
			document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/`;
		},
	}
});
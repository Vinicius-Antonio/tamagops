export namespace collector {
	
	export class TopProcess {
	    pid: number;
	    name: string;
	    cpu_percent: number;
	
	    static createFrom(source: any = {}) {
	        return new TopProcess(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.cpu_percent = source["cpu_percent"];
	    }
	}
	export class SystemSnapshot {
	    cpu_percent: number;
	    cpu_temp_celsius: number;
	    board_temp_celsius: number;
	    sensors_available: boolean;
	    ram_total_bytes: number;
	    ram_used_bytes: number;
	    ram_available_bytes: number;
	    ram_used_percent: number;
	    swap_total_bytes: number;
	    swap_used_bytes: number;
	    swap_used_percent: number;
	    disk_total_bytes: number;
	    disk_used_bytes: number;
	    disk_free_bytes: number;
	    disk_used_percent: number;
	    load_avg_1min: number;
	    load_avg_5min: number;
	    load_avg_15min: number;
	    load_available: boolean;
	    uptime_seconds: number;
	    top_process: TopProcess;
	    // Go type: time
	    collected_at: any;
	
	    static createFrom(source: any = {}) {
	        return new SystemSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu_percent = source["cpu_percent"];
	        this.cpu_temp_celsius = source["cpu_temp_celsius"];
	        this.board_temp_celsius = source["board_temp_celsius"];
	        this.sensors_available = source["sensors_available"];
	        this.ram_total_bytes = source["ram_total_bytes"];
	        this.ram_used_bytes = source["ram_used_bytes"];
	        this.ram_available_bytes = source["ram_available_bytes"];
	        this.ram_used_percent = source["ram_used_percent"];
	        this.swap_total_bytes = source["swap_total_bytes"];
	        this.swap_used_bytes = source["swap_used_bytes"];
	        this.swap_used_percent = source["swap_used_percent"];
	        this.disk_total_bytes = source["disk_total_bytes"];
	        this.disk_used_bytes = source["disk_used_bytes"];
	        this.disk_free_bytes = source["disk_free_bytes"];
	        this.disk_used_percent = source["disk_used_percent"];
	        this.load_avg_1min = source["load_avg_1min"];
	        this.load_avg_5min = source["load_avg_5min"];
	        this.load_avg_15min = source["load_avg_15min"];
	        this.load_available = source["load_available"];
	        this.uptime_seconds = source["uptime_seconds"];
	        this.top_process = this.convertValues(source["top_process"], TopProcess);
	        this.collected_at = this.convertValues(source["collected_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace engine {
	
	export class PetState {
	    name: string;
	    level: number;
	    current_xp: number;
	    next_level_xp: number;
	    hp: number;
	    max_hp: number;
	    mood: string;
	    active_debuffs: string[];
	
	    static createFrom(source: any = {}) {
	        return new PetState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.level = source["level"];
	        this.current_xp = source["current_xp"];
	        this.next_level_xp = source["next_level_xp"];
	        this.hp = source["hp"];
	        this.max_hp = source["max_hp"];
	        this.mood = source["mood"];
	        this.active_debuffs = source["active_debuffs"];
	    }
	}

}

export namespace main {
	
	export class DashboardPayload {
	    timestamp: string;
	    pet: engine.PetState;
	    hardware: collector.SystemSnapshot;
	    recent_log: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.pet = this.convertValues(source["pet"], engine.PetState);
	        this.hardware = this.convertValues(source["hardware"], collector.SystemSnapshot);
	        this.recent_log = source["recent_log"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}


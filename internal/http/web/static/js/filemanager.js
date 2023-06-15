class FileManager {
    constructor(baseURL) {
        this.baseURL = baseURL;
    }

    async getDir(id = '') {
        const response = await fetch(`${this.baseURL}/api/v1/directories/${id}`);
        return response.json();
    }

    async createDir(directory) {
        const response = await fetch(`${this.baseURL}/api/v1/directories/`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(directory),
        });
        return response.json();
    }

    async renameDir(id, name) {
        const response = await fetch(`${this.baseURL}/api/v1/directories/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({name: name}),
        });
        return response.json();
    }

    async delDir(id) {
        const response = await fetch(`${this.baseURL}/api/v1/directories/${id}`, {
            method: 'DELETE',
        });
        return response.json();
    }

    async createFile(directoryId, file) {
        const formData = new FormData();
        formData.append('file', file);

        const response = await fetch(`${this.baseURL}/api/v1/directories/${directoryId}/files`, {
            method: 'POST',
            body: formData,
        });
        return response.json();
    }

    async renameFile(directoryId, fileId, name) {
        const response = await fetch(`${this.baseURL}/api/v1/directories/${directoryId}/files/${fileId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({name: name}),
        });
        return response.json();
    }

    async deleteFile(directoryId, fileId) {
        const response = await fetch(`${this.baseURL}/api/v1/directories/${directoryId}/files/${fileId}`, {
            method: 'DELETE',
        });
        return response.json();
    }
}

/**
 * K8s API Service
 * Handle all network requests here.
 */
export const k8sService = {
  //   async fetchTableData(tab, refresh = false) {
  //     const endpoint = tab === 'sa' ? '/api/table' : '/api/groups';
  //     // Append the refresh query parameter
  //     const url = `${endpoint}?refresh=${refresh}`;
  //     const res = await fetch(url);
  //     if (!res.ok) throw new Error(`API Error: ${res.status}`);
  //     return await res.json();
  //   },
async fetchTableData(tab, refresh = false) {
    const endpoint = tab === 'cilium' ? '/api/cilium' : (tab === 'groups' ? '/api/groups' : '/api/table');
    const url = refresh ? `${endpoint}?refresh=true` : endpoint;
    const res = await fetch(url);
    return await res.json();
},

  downloadExport(data, tabName) {
    const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(data, null, 2));
    const link = document.createElement('a');
    link.setAttribute("href", dataStr);
    link.setAttribute("download", `k8s_audit_${tabName}.json`);
    link.click();
  }
};
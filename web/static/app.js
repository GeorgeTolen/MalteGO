// MalteGO Web UI — Cytoscape.js graph

const ENTITY_COLORS = {
  'maltego.IPv4Address':   '#58a6ff',
  'maltego.Person':        '#f78166',
  'maltego.Organization':  '#d2a8ff',
  'maltego.AS':            '#ffa657',
  'maltego.Port':          '#56d364',
  'maltego.CVE':           '#f85149',
  'maltego.Hashtag':       '#79c0ff',
  'greynoise.noise':       '#ffa657',
  'greynoise.classification': '#e6edf3',
  'default':               '#8b949e',
};

function entityColor(type) {
  return ENTITY_COLORS[type] || ENTITY_COLORS.default;
}

function entityLabel(type) {
  return type.split('.').pop();
}

// ── Init Cytoscape ──────────────────────────────────────────────────────────
const cy = cytoscape({
  container: document.getElementById('cy'),
  style: [
    {
      selector: 'node',
      style: {
        'background-color': 'data(color)',
        'label':            'data(label)',
        'color':            '#e6edf3',
        'text-valign':      'bottom',
        'text-halign':      'center',
        'font-size':        '11px',
        'text-margin-y':    '4px',
        'width':            '36px',
        'height':           '36px',
        'border-width':     '2px',
        'border-color':     '#30363d',
      }
    },
    {
      selector: 'node:selected',
      style: { 'border-color': '#f78166', 'border-width': '3px' }
    },
    {
      selector: 'node.source',
      style: { 'width': '48px', 'height': '48px', 'border-color': '#58a6ff', 'border-width': '3px' }
    },
    {
      selector: 'edge',
      style: {
        'width':          '1.5px',
        'line-color':     '#30363d',
        'target-arrow-color': '#30363d',
        'target-arrow-shape': 'triangle',
        'curve-style':    'bezier',
        'label':          'data(label)',
        'font-size':      '10px',
        'color':          '#8b949e',
        'text-rotation':  'autorotate',
      }
    }
  ],
  layout: { name: 'cose' },
  userZoomingEnabled: true,
  userPanningEnabled: true,
  boxSelectionEnabled: false,
});

// ── Node click → detail panel ───────────────────────────────────────────────
cy.on('tap', 'node', function(e) {
  const data = e.target.data();
  document.getElementById('detail-panel').style.display = 'flex';
  document.getElementById('detail-type').textContent  = data.entityType || data.type || '';
  document.getElementById('detail-value').textContent = data.value || data.id;

  const propsEl = document.getElementById('detail-props');
  propsEl.innerHTML = '';
  const props = data.properties || {};
  Object.entries(props).forEach(([k, v]) => {
    if (!v) return;
    const row = document.createElement('div');
    row.className = 'prop-row';
    row.innerHTML = `<span class="prop-key">${k}</span><span class="prop-val">${v}</span>`;
    propsEl.appendChild(row);
  });
});

cy.on('tap', function(e) {
  if (e.target === cy) {
    document.getElementById('detail-panel').style.display = 'none';
  }
});

// ── API Key ─────────────────────────────────────────────────────────────────
function getAPIKey() {
  return localStorage.getItem('greynoise_api_key') || '';
}

function initAPIKey() {
  const saved = getAPIKey();
  if (saved) {
    document.getElementById('input-apikey').value = saved;
    document.getElementById('apikey-status').textContent = 'API key set';
  }
}

function saveAPIKey() {
  const key = document.getElementById('input-apikey').value.trim();
  if (key) {
    localStorage.setItem('greynoise_api_key', key);
    document.getElementById('apikey-status').textContent = 'API key saved';
  } else {
    localStorage.removeItem('greynoise_api_key');
    document.getElementById('apikey-status').textContent = 'API key cleared';
  }
}

// ── Load transforms list ────────────────────────────────────────────────────
async function loadTransforms() {
  try {
    const res = await fetch('/api/transforms');
    const data = await res.json();
    const sel = document.getElementById('input-transform');
    sel.innerHTML = '';
    (data.transforms || []).forEach(name => {
      const opt = document.createElement('option');
      opt.value = name;
      opt.textContent = name;
      sel.appendChild(opt);
    });
  } catch (err) {
    setStatus('Failed to load transforms: ' + err.message, 'error');
  }
}

// ── Run transform ───────────────────────────────────────────────────────────
async function runTransform() {
  const value     = document.getElementById('input-value').value.trim();
  const transform = document.getElementById('input-transform').value;

  if (!value)     { setStatus('Enter an entity value', 'error'); return; }
  if (!transform) { setStatus('Select a transform', 'error'); return; }

  setStatus('Running ' + transform + '…', '');
  document.getElementById('btn-run').disabled = true;

  try {
    const body = { value, entity_type: 'maltego.IPv4Address' };
    const apiKey = getAPIKey();
    if (apiKey) body.api_key = apiKey;

    const res = await fetch('/api/run/' + transform, {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify(body),
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || res.statusText);
    }

    const data = await res.json();
    renderGraph(value, transform, data.entities || [], data.messages || []);

    const count = (data.entities || []).length;
    setStatus(`✓ ${count} entit${count === 1 ? 'y' : 'ies'} returned`, 'ok');
  } catch (err) {
    setStatus('Error: ' + err.message, 'error');
  } finally {
    document.getElementById('btn-run').disabled = false;
  }
}

function renderGraph(sourceValue, transform, entities, messages) {
  // Find or create source node
  let sourceId = 'src-' + sourceValue;
  if (!cy.getElementById(sourceId).length) {
    cy.add({
      group: 'nodes',
      data: {
        id:         sourceId,
        label:      sourceValue,
        value:      sourceValue,
        entityType: 'maltego.IPv4Address',
        color:      entityColor('maltego.IPv4Address'),
        properties: {},
      },
      classes: 'source',
    });
  }

  entities.forEach(entity => {
    if (entity.value === sourceValue) return; // skip self

    const nodeId = entity.type + '::' + entity.value;
    if (!cy.getElementById(nodeId).length) {
      cy.add({
        group: 'nodes',
        data: {
          id:         nodeId,
          label:      entity.value.length > 18 ? entity.value.slice(0, 16) + '…' : entity.value,
          value:      entity.value,
          entityType: entity.type,
          color:      entityColor(entity.type),
          properties: entity.properties || {},
        },
      });
    }

    const edgeId = 'e-' + sourceId + '-' + nodeId;
    if (!cy.getElementById(edgeId).length) {
      cy.add({
        group: 'edges',
        data: {
          id:     edgeId,
          source: sourceId,
          target: nodeId,
          label:  entityLabel(entity.type),
        },
      });
    }
  });

  // Re-run layout
  cy.layout({
    name:         'cose',
    animate:      true,
    animationDuration: 400,
    randomize:    false,
    nodeRepulsion: 8000,
    idealEdgeLength: 100,
  }).run();
}

function clearGraph() {
  cy.elements().remove();
  document.getElementById('detail-panel').style.display = 'none';
  setStatus('Graph cleared', '');
}

// ── Export ───────────────────────────────────────────────────────────────────
function exportPNG() {
  if (!cy.elements().length) { setStatus('Graph is empty', 'error'); return; }
  const png = cy.png({ full: true, scale: 2, bg: '#0d1117' });
  const a = document.createElement('a');
  a.href = png;
  a.download = 'maltego-graph.png';
  a.click();
}

function exportJSON() {
  if (!cy.elements().length) { setStatus('Graph is empty', 'error'); return; }
  const json = JSON.stringify(cy.json(), null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = 'maltego-graph.json';
  a.click();
  URL.revokeObjectURL(a.href);
}

function setStatus(msg, cls) {
  const el = document.getElementById('status');
  el.textContent  = msg;
  el.className    = cls || '';
}

// ── Save / Load graphs ───────────────────────────────────────────────────────

let graphsOffset = 0;
const GRAPHS_LIMIT = 10;

async function loadSavedGraphs(reset = true) {
  if (reset) graphsOffset = 0;
  try {
    const res = await fetch(`/api/graphs?limit=${GRAPHS_LIMIT}&offset=${graphsOffset}`);
    const data = await res.json();
    renderSavedGraphs(data.graphs || [], data.total || 0, reset);
    graphsOffset += (data.graphs || []).length;
  } catch (_) {}
}

function renderSavedGraphs(graphs, total, reset) {
  const list = document.getElementById('saved-graphs-list');
  if (reset) list.innerHTML = '';

  if (reset && !graphs.length) {
    list.innerHTML = '<div style="font-size:11px;color:var(--muted);text-align:center;padding:4px">No saved graphs</div>';
    document.getElementById('btn-load-more').style.display = 'none';
    return;
  }

  graphs.forEach(g => {
    const item = document.createElement('div');
    item.className = 'graph-item';
    const date = new Date(g.updated_at).toLocaleDateString();
    item.innerHTML = `
      <span class="graph-item-name" title="${g.name}">${g.name}</span>
      <span class="graph-item-date">${date}</span>
      <button class="graph-item-rename" title="Rename" data-id="${g.id}">✎</button>
      <button class="graph-item-del" title="Delete" data-id="${g.id}">×</button>
    `;
    item.querySelector('.graph-item-name').addEventListener('click', () => openGraph(g.id));
    item.querySelector('.graph-item-rename').addEventListener('click', e => {
      e.stopPropagation();
      startRename(item, g.id, g.name);
    });
    item.querySelector('.graph-item-del').addEventListener('click', e => {
      e.stopPropagation();
      deleteGraph(g.id);
    });
    list.appendChild(item);
  });

  const loadMore = document.getElementById('btn-load-more');
  loadMore.style.display = graphsOffset < total ? 'block' : 'none';
}

function startRename(item, id, currentName) {
  const nameEl = item.querySelector('.graph-item-name');
  const input = document.createElement('input');
  input.value = currentName;
  input.className = 'graph-item-name';
  input.style.cursor = 'text';
  nameEl.replaceWith(input);
  input.focus();
  input.select();

  const commit = async () => {
    const newName = input.value.trim();
    if (newName && newName !== currentName) {
      await renameGraph(id, newName);
    } else {
      loadSavedGraphs();
    }
  };
  input.addEventListener('blur', commit);
  input.addEventListener('keydown', e => {
    if (e.key === 'Enter') { e.preventDefault(); input.blur(); }
    if (e.key === 'Escape') { input.removeEventListener('blur', commit); loadSavedGraphs(); }
  });
}

async function saveGraph() {
  if (!cy.elements().length) {
    setStatus('Nothing to save — graph is empty', 'error');
    return;
  }
  const nameInput = document.getElementById('graph-name');
  const name = nameInput.value.trim() || 'Graph ' + new Date().toLocaleString();
  const data = JSON.stringify(cy.json());

  try {
    const res = await fetch('/api/graphs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, data }),
    });
    if (!res.ok) {
      const ct = res.headers.get('content-type') || '';
      const msg = ct.includes('json')
        ? (await res.json()).error
        : await res.text();
      if (res.status === 404) throw new Error('Database not available — start Docker first');
      throw new Error(msg || res.statusText);
    }
    nameInput.value = '';
    setStatus('Graph saved: ' + name, 'ok');
    loadSavedGraphs();
  } catch (err) {
    setStatus('Save failed: ' + err.message, 'error');
  }
}

async function openGraph(id) {
  try {
    const res = await fetch('/api/graphs/' + id);
    if (!res.ok) {
      if (res.status === 404) throw new Error('Graph not found');
      throw new Error(res.statusText);
    }
    const g = await res.json();
    cy.elements().remove();
    cy.json(JSON.parse(g.data));
    cy.layout({ name: 'preset' }).run();
    setStatus('Loaded: ' + g.name, 'ok');
  } catch (err) {
    setStatus('Load failed: ' + err.message, 'error');
  }
}

async function renameGraph(id, name) {
  try {
    await fetch(`/api/graphs/${id}/rename`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    });
    loadSavedGraphs();
  } catch (err) {
    setStatus('Rename failed: ' + err.message, 'error');
  }
}

async function deleteGraph(id) {
  try {
    await fetch('/api/graphs/' + id, { method: 'DELETE' });
    loadSavedGraphs();
  } catch (err) {
    setStatus('Delete failed: ' + err.message, 'error');
  }
}

// ── Wire up buttons ─────────────────────────────────────────────────────────
document.getElementById('btn-run').addEventListener('click', runTransform);
document.getElementById('btn-clear').addEventListener('click', clearGraph);
document.getElementById('btn-save').addEventListener('click', saveGraph);
document.getElementById('btn-export-png').addEventListener('click', exportPNG);
document.getElementById('btn-export-json').addEventListener('click', exportJSON);
document.getElementById('btn-apikey-save').addEventListener('click', saveAPIKey);
document.getElementById('btn-load-more').addEventListener('click', () => loadSavedGraphs(false));
document.getElementById('input-value').addEventListener('keydown', e => {
  if (e.key === 'Enter') runTransform();
});

// ── Boot ─────────────────────────────────────────────────────────────────────
initAPIKey();
loadTransforms();
loadSavedGraphs();

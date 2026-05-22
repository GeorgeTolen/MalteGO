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
let nodeCounter = 0;
function uid() { return 'n' + (++nodeCounter); }

async function runTransform() {
  const value     = document.getElementById('input-value').value.trim();
  const transform = document.getElementById('input-transform').value;

  if (!value)     { setStatus('Enter an entity value', 'error'); return; }
  if (!transform) { setStatus('Select a transform', 'error'); return; }

  setStatus('Running ' + transform + '…', '');
  document.getElementById('btn-run').disabled = true;

  try {
    const res = await fetch('/api/run/' + transform, {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({ value, entity_type: 'maltego.IPv4Address' }),
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

function setStatus(msg, cls) {
  const el = document.getElementById('status');
  el.textContent  = msg;
  el.className    = cls || '';
}

// ── Wire up buttons ─────────────────────────────────────────────────────────
document.getElementById('btn-run').addEventListener('click', runTransform);
document.getElementById('btn-clear').addEventListener('click', clearGraph);
document.getElementById('input-value').addEventListener('keydown', e => {
  if (e.key === 'Enter') runTransform();
});

// ── Boot ─────────────────────────────────────────────────────────────────────
loadTransforms();

window.renderKeystrokeReport = function (config) {
    const { data, containerId, legendId, noDataId } = config;
    const container = document.getElementById(containerId);
    const legendContainer = document.getElementById(legendId);
    const noDataMsg = document.getElementById(noDataId);

    if (!container) return; // Should not happen if DOM matches

    console.log("Rendering report for", containerId, "Data points:", data?.points?.length);

    if (!data || !data.points || data.points.length === 0) {
        if (noDataMsg) noDataMsg.classList.remove('hidden');
        container.innerHTML = '';
        if (legendContainer) legendContainer.innerHTML = '';
        return;
    }

    const results = data.points;

    // SVG dimensions
    const width = 800;
    const height = 300;
    const padding = { top: 20, right: 20, bottom: 40, left: 50 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    // Y-axis: 0 to 5
    const maxY = 5;
    const minY = 0;

    // Create SVG
    let html = `<svg width="${width}" height="${height}" style="background: white; border: 1px solid #ddd; border-radius: 6px;">`;

    // Draw Y-axis
    html += `<line x1="${padding.left}" y1="${padding.top}" x2="${padding.left}" y2="${height - padding.bottom}" stroke="#333" stroke-width="2"/>`;

    // Draw X-axis
    html += `<line x1="${padding.left}" y1="${height - padding.bottom}" x2="${width - padding.right}" y2="${height - padding.bottom}" stroke="#333" stroke-width="2"/>`;

    // Y-axis labels and grid lines
    for (let i = 0; i <= 5; i++) {
        const y = padding.top + (chartHeight * (5 - i) / 5);
        html += `<text x="${padding.left - 10}" y="${y + 5}" text-anchor="end" font-size="12" fill="#666">${i}</text>`;
        html += `<line x1="${padding.left}" y1="${y}" x2="${width - padding.right}" y2="${y}" stroke="#eee" stroke-width="1"/>`;
    }

    // Threshold line at 1.2 (legit vs suspicious)
    const thresholdY = padding.top + (chartHeight * (1 - 1.2 / maxY));
    html += `<line x1="${padding.left}" y1="${thresholdY}" x2="${width - padding.right}" y2="${thresholdY}" stroke="#ff9800" stroke-width="2" stroke-dasharray="5,5"/>`;
    html += `<text x="${width - padding.right + 5}" y="${thresholdY + 5}" font-size="11" fill="#ff9800" font-weight="bold">1.2</text>`;

    // Threshold labels
    html += `<text x="${padding.left + 10}" y="${padding.top + 15}" font-size="10" fill="#d32f2f" font-weight="bold">Sospechoso</text>`;
    html += `<text x="${padding.left + 10}" y="${height - padding.bottom - 10}" font-size="10" fill="#388e3c" font-weight="bold">Legítimo</text>`;


    // Y-axis label
    html += `<text x="15" y="${height / 2}" text-anchor="middle" font-size="12" fill="#333" transform="rotate(-90, 15, ${height / 2})">SMD</text>`;

    // Calculate X positions for each window
    const windowCount = results.length;
    const xStep = chartWidth / (windowCount > 1 ? windowCount - 1 : 1);

    // Draw line and points
    let points = [];

    for (let i = 0; i < results.length; i++) {
        const result = results[i];
        const x = padding.left + (i * xStep);
        const isInactive = result.isInactive;

        // Cap SMD at 5
        const smdValue = Math.min(result.smd, 5);
        const y = padding.top + (chartHeight * (1 - smdValue / maxY));

        points.push({ x, y, result, isInactive, index: i });
    }

    // Draw line segments (break at inactive windows)
    let segmentStart = null;
    for (let i = 0; i < points.length; i++) {
        const point = points[i];

        if (!point.isInactive) {
            if (segmentStart === null) {
                segmentStart = i;
            }
        } else {
            // Draw segment if we have one
            if (segmentStart !== null && i - segmentStart > 1) {
                let pathData = '';
                for (let j = segmentStart; j < i; j++) {
                    const p = points[j];
                    if (pathData === '') {
                        pathData = `M ${p.x} ${p.y}`;
                    } else {
                        pathData += ` L ${p.x} ${p.y}`;
                    }
                }
                html += `<path d="${pathData}" fill="none" stroke="#007bff" stroke-width="2"/>`;
            }
            segmentStart = null;
        }
    }

    // Draw final segment if exists
    if (segmentStart !== null && points.length - segmentStart > 1) {
        let pathData = '';
        for (let j = segmentStart; j < points.length; j++) {
            const p = points[j];
            if (!p.isInactive) {
                if (pathData === '') {
                    pathData = `M ${p.x} ${p.y}`;
                } else {
                    pathData += ` L ${p.x} ${p.y}`;
                }
            }
        }
        if (pathData !== '') {
            html += `<path d="${pathData}" fill="none" stroke="#007bff" stroke-width="2"/>`;
        }
    }

    // Draw points
    for (const point of points) {
        const { x, y, result, isInactive, index } = point;

        if (isInactive) {
            // Inactive time: larger gray X marker with background
            const xSize = 8;
            const yPos = height - padding.bottom - 15;

            // Group for hover
            html += `<g>`;
            html += `<title>Ventana ${index + 1}: ${result.strokeCount} teclas (Inactivo)</title>`;

            // Background circle
            html += `<circle cx="${x}" cy="${yPos}" r="12" fill="#f0f0f0" stroke="#999" stroke-width="1"/>`;

            // X marker
            html += `<line x1="${x - xSize}" y1="${yPos - xSize}" x2="${x + xSize}" y2="${yPos + xSize}" stroke="#666" stroke-width="2.5"/>`;
            html += `<line x1="${x - xSize}" y1="${yPos + xSize}" x2="${x + xSize}" y2="${yPos - xSize}" stroke="#666" stroke-width="2.5"/>`;
            html += `</g>`;
        } else if (result.isProfile) {
            // Profile: green square
            html += `<g>`;
            html += `<title>Ventana ${index + 1}: ${result.strokeCount} teclas (Perfil) - SMD: ${result.smd.toFixed(2)}</title>`;
            html += `<rect x="${x - 6}" y="${y - 6}" width="12" height="12" fill="#28a745" stroke="#1e7e34" stroke-width="2"/>`;
            html += `</g>`;
        } else {
            // Regular point: color based on threshold
            const isSuspicious = result.smd > 1.2;
            const color = isSuspicious ? '#d32f2f' : '#388e3c';
            const strokeColor = isSuspicious ? '#b71c1c' : '#2e7d32';
            const status = isSuspicious ? 'Sospechoso' : 'Legítimo';

            html += `<g>`;
            html += `<title>Ventana ${index + 1}: ${result.strokeCount} teclas (${status}) - SMD: ${result.smd.toFixed(2)}</title>`;
            html += `<circle cx="${x}" cy="${y}" r="5" fill="${color}" stroke="${strokeColor}" stroke-width="2"/>`;
            html += `</g>`;
        }

        // X-axis label
        html += `<text x="${x}" y="${height - padding.bottom + 20}" text-anchor="middle" font-size="11" fill="#666">W${index + 1}</text>`;

        // Hover info (simplified - just show value on top of point if not inactive)
        if (!isInactive) {
            html += `<text x="${x}" y="${y - 10}" text-anchor="middle" font-size="10" fill="#333">${result.smd.toFixed(2)}</text>`;
        }
    }

    html += '</svg>';

    // Render chart
    container.innerHTML = html;

    // Render legend
    let legendHtml = '<div style="display: flex; gap: 1.5rem; flex-wrap: wrap;">';
    legendHtml += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #28a745; border: 2px solid #1e7e34; margin-right: 0.3rem;"></span> Perfil</div>';
    legendHtml += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #388e3c; border: 2px solid #2e7d32; border-radius: 50%; margin-right: 0.3rem;"></span> Legítimo (≤1.2)</div>';
    legendHtml += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #d32f2f; border: 2px solid #b71c1c; border-radius: 50%; margin-right: 0.3rem;"></span> Sospechoso (&gt;1.2)</div>';
    legendHtml += '<div><span style="display: inline-block; margin-right: 0.3rem;">✕</span> Inactivo (&lt;10 teclas)</div>';
    legendHtml += '</div>';

    // Find profile index for legend extra info
    const profileIdx = results.findIndex(r => r.isProfile);
    if (profileIdx !== -1) {
        legendHtml += `<p style="margin-top: 0.5rem; color: #9ca3af;">Perfil: Ventana ${profileIdx + 1} (${results[profileIdx].strokeCount} teclas)</p>`;
    }

    if (legendContainer) legendContainer.innerHTML = legendHtml;

};

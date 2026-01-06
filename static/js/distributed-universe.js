class DistributedUniverse {
    constructor() {
        this.canvas = document.getElementById('universe-canvas');
        this.ctx = this.canvas.getContext('2d');
        this.concepts = [];
        this.particles = [];
        this.connections = [];
        this.activeConcept = null;
        this.mouse = { x: 0, y: 0 };
        this.time = 0;

        this.init();
        this.createConcepts();
        this.createParticles();
        this.setupEventListeners();
        this.animate();
    }

    init() {
        this.resizeCanvas();
        window.addEventListener('resize', () => this.resizeCanvas());

        // Set initial controls
        document.getElementById('speedControl').value = 1;
        document.getElementById('particleControl').value = 50;
    }

    resizeCanvas() {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
    }

    createConcepts() {
        this.concepts = [
            {
                id: 'raft',
                name: 'Raft Consensus',
                category: 'consensus',
                icon: '👑',
                description: `A consensus algorithm designed as an alternative to Paxos. It's easier to understand and implement while providing the same safety guarantees.`,
                details: `
                    <h4>Core Principles</h4>
                    <ul>
                        <li><strong>Leader Election:</strong> Nodes elect a single leader</li>
                        <li><strong>Log Replication:</strong> Leader replicates logs to followers</li>
                        <li><strong>Safety:</strong> Committed entries are never overwritten</li>
                    </ul>
                    
                    <h4>Node States</h4>
                    <ul>
                        <li><strong>Leader:</strong> Handles all client requests</li>
                        <li><strong>Follower:</strong> Passive, responds to RPCs</li>
                        <li><strong>Candidate:</strong> Intermediate state during elections</li>
                    </ul>
                    
                    <h4>Use Cases</h4>
                    <ul>
                        <li>etcd - distributed key-value store</li>
                        <li>Consul - service discovery</li>
                        <li>Kubernetes - cluster management</li>
                    </ul>
                `,
                difficulty: 4,
                related: ['paxos', 'consensus', 'leader-election']
            },
            {
                id: 'paxos',
                name: 'Paxos',
                category: 'consensus',
                icon: '⚖️',
                description: `The classic consensus protocol that ensures safety in asynchronous systems with failure detectors.`,
                details: `Paxos guarantees consensus among distributed processes even when some processes fail.`,
                difficulty: 5,
                related: ['raft', 'consensus', 'byzantine']
            },
            {
                id: 'kafka',
                name: 'Apache Kafka',
                category: 'messaging',
                icon: '📨',
                description: `Distributed event streaming platform capable of handling trillions of events a day.`,
                details: `Kafka provides high-throughput, low-latency platform for handling real-time data feeds.`,
                difficulty: 3,
                related: ['message-queue', 'stream-processing', 'pub-sub']
            },
            {
                id: 'consul',
                name: 'Consul',
                category: 'coordination',
                icon: '🔗',
                description: `Service networking solution to connect and secure services across any runtime platform.`,
                details: `Consul provides service discovery, configuration, and segmentation functionality.`,
                difficulty: 3,
                related: ['service-discovery', 'configuration', 'health-check']
            },
            {
                id: 'zookeeper',
                name: 'ZooKeeper',
                category: 'coordination',
                icon: '🐘',
                description: `Centralized service for maintaining configuration information, naming, and synchronization.`,
                details: `ZooKeeper is used by distributed systems for coordination and consensus.`,
                difficulty: 4,
                related: ['coordination', 'configuration', 'consensus']
            },
            {
                id: 'cassandra',
                name: 'Apache Cassandra',
                category: 'storage',
                icon: '🗃️',
                description: `Highly scalable distributed NoSQL database designed to handle large amounts of data across many commodity servers.`,
                details: `Cassandra provides high availability with no single point of failure.`,
                difficulty: 4,
                related: ['database', 'replication', 'partitioning']
            },
            {
                id: 'grpc',
                name: 'gRPC',
                category: 'messaging',
                icon: '🚀',
                description: `High-performance, open-source universal RPC framework developed by Google.`,
                details: `gRPC uses HTTP/2 for transport, Protocol Buffers as interface description language.`,
                difficulty: 3,
                related: ['rpc', 'protocol-buffers', 'http2']
            },
            {
                id: 'kubernetes',
                name: 'Kubernetes',
                category: 'architecture',
                icon: '☸️',
                description: `Portable, extensible, open-source platform for managing containerized workloads and services.`,
                details: `Kubernetes automates deployment, scaling, and management of containerized applications.`,
                difficulty: 5,
                related: ['orchestration', 'containers', 'microservices']
            },
            {
                id: 'circuit-breaker',
                name: 'Circuit Breaker',
                category: 'patterns',
                icon: '⚡',
                description: `Design pattern used to detect failures and encapsulate the logic of preventing a failure from constantly recurring.`,
                details: `Prevents cascading failures in distributed systems.`,
                difficulty: 2,
                related: ['resilience', 'failure-detection', 'microservices']
            },
            {
                id: 'consistent-hashing',
                name: 'Consistent Hashing',
                category: 'patterns',
                icon: '🎯',
                description: `Special kind of hashing that minimizes reorganization when nodes are added or removed.`,
                details: `Used in distributed hash tables and load balancers.`,
                difficulty: 3,
                related: ['hashing', 'load-balancing', 'partitioning']
            },
            {
                id: 'vector-clocks',
                name: 'Vector Clocks',
                category: 'consensus',
                icon: '🕰️',
                description: `Algorithm for generating a partial ordering of events in a distributed system.`,
                details: `Used for tracking causality between events in distributed systems.`,
                difficulty: 4,
                related: ['clocks', 'causality', 'event-ordering']
            },
            {
                id: 'byzantine',
                name: 'Byzantine Fault',
                category: 'consensus',
                icon: '👑',
                description: `The Byzantine Generals Problem describes the difficulty decentralized systems have in agreeing on a strategy.`,
                details: `A node may fail in arbitrary ways, including sending contradictory messages.`,
                difficulty: 5,
                related: ['fault-tolerance', 'consensus', 'paxos']
            }
        ];

        // Position concepts in 3D space
        this.positionConcepts();
        this.createHTMLNodes();
    }

    positionConcepts() {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;

        // Calculate available space
        const padding = 100;
        const minX = padding;
        const maxX = this.canvas.width - padding;
        const minY = padding;
        const maxY = this.canvas.height - padding;

        // Group concepts by category for better organization
        const categories = {};
        this.concepts.forEach(concept => {
            if (!categories[concept.category]) {
                categories[concept.category] = [];
            }
            categories[concept.category].push(concept);
        });

        // Position each category in its own region
        const categoryPositions = {
            'consensus': { x: centerX * 0.3, y: centerY * 0.4 },
            'messaging': { x: centerX * 1.7, y: centerY * 0.4 },
            'coordination': { x: centerX * 0.3, y: centerY * 1.6 },
            'storage': { x: centerX * 1.7, y: centerY * 1.6 },
            'architecture': { x: centerX, y: centerY * 0.2 },
            'patterns': { x: centerX, y: centerY * 1.8 }
        };

        let conceptIndex = 0;

        // Position concepts with proper scattering
        Object.keys(categories).forEach(category => {
            const concepts = categories[category];
            const basePos = categoryPositions[category] || { x: centerX, y: centerY };

            // Calculate spacing for this cluster
            const clusterRadius = Math.min(150, concepts.length * 40);
            const angleStep = (Math.PI * 2) / concepts.length;

            concepts.forEach((concept, index) => {
                // Create a spiral pattern within the cluster
                const spiralAngle = angleStep * index;
                const spiralRadius = clusterRadius * (index / concepts.length);

                // Calculate position within cluster
                let x = basePos.x + Math.cos(spiralAngle) * spiralRadius;
                let y = basePos.y + Math.sin(spiralAngle) * spiralRadius;

                // Add some randomness for natural scattering
                x += (Math.random() - 0.5) * 80;
                y += (Math.random() - 0.5) * 80;

                // Ensure positions stay within bounds
                x = Math.max(minX, Math.min(maxX, x));
                y = Math.max(minY, Math.min(maxY, y));

                // Set depth for 3D effect
                const z = (Math.random() - 0.5) * 200;

                // Assign properties
                concept.x = x;
                concept.y = y;
                concept.z = z;
                concept.rotation = 0;
                concept.speed = 0.5 + Math.random() * 0.5;
                concept.baseX = x; // Store original position for animation
                concept.baseY = y;
                concept.baseZ = z;

                // Add some offset for floating animation
                concept.floatOffset = Math.random() * Math.PI * 2;
                concept.floatAmplitude = 10 + Math.random() * 20;
                concept.floatSpeed = 0.001 + Math.random() * 0.002;

                conceptIndex++;
            });
        });

        // Create connections after positioning
        setTimeout(() => this.createConnections(), 100);
    }

    createHTMLNodes() {
        const container = document.querySelector('.universe-container');

        this.concepts.forEach(concept => {
            const node = document.createElement('div');
            node.className = 'concept-node pulse float';
            node.dataset.id = concept.id;
            node.dataset.category = concept.category;

            node.innerHTML = `
                <div class="concept-icon">${concept.icon}</div>
                <div class="concept-label">${concept.name}</div>
            `;

            node.style.left = `${concept.x}px`;
            node.style.top = `${concept.y}px`;

            node.addEventListener('click', (e) => {
                e.stopPropagation();
                this.showConceptDetails(concept);
            });

            container.appendChild(node);
            concept.element = node;
        });
    }

    createParticles() {
        const count = parseInt(document.getElementById('particleControl').value);
        this.particles = [];

        for (let i = 0; i < count; i++) {
            this.particles.push({
                x: Math.random() * this.canvas.width,
                y: Math.random() * this.canvas.height,
                z: Math.random() * 200 - 100,
                speed: 0.2 + Math.random() * 0.3,
                size: Math.random() * 2 + 1,
                opacity: Math.random() * 0.5 + 0.1
            });
        }
    }

    createConnections() {
        this.connections = [];

        // Create connections between related concepts
        this.concepts.forEach((concept1, i) => {
            concept1.related.forEach(relatedId => {
                const concept2 = this.concepts.find(c => c.id === relatedId);
                if (concept2) {
                    this.connections.push({
                        from: concept1,
                        to: concept2,
                        strength: 0.3
                    });
                }
            });
        });
    }

    setupEventListeners() {
        // Mouse movement
        document.addEventListener('mousemove', (e) => {
            this.mouse.x = e.clientX;
            this.mouse.y = e.clientY;
        });

        // Control listeners
        document.getElementById('speedControl').addEventListener('input', (e) => {
            const speed = parseFloat(e.target.value);
            this.concepts.forEach(c => c.speed = speed * (0.5 + Math.random() * 0.5));
        });

        document.getElementById('particleControl').addEventListener('input', (e) => {
            this.createParticles();
        });

        document.getElementById('showConnections').addEventListener('change', (e) => {
            const show = e.target.checked;
            document.querySelectorAll('.connection-line').forEach(el => {
                el.style.display = show ? 'block' : 'none';
            });
        });

        // Filter buttons
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');

                const filter = btn.dataset.filter;
                this.filterConcepts(filter);
            });
        });

        // Close knowledge panel
        document.getElementById('closeKnowledge').addEventListener('click', () => {
            this.hideConceptDetails();
        });

        // Close on escape
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.hideConceptDetails();
                document.getElementById('universeControls').classList.remove('visible');
            }
        });

        // Tour button
        document.getElementById('tourBtn').addEventListener('click', () => {
            this.startGuidedTour();
        });

        // Quiz button
        document.getElementById('quizBtn').addEventListener('click', () => {
            this.startQuiz();
        });
    }

    filterConcepts(filter) {
        this.concepts.forEach(concept => {
            const element = concept.element;
            if (filter === 'all' || concept.category === filter) {
                element.style.opacity = '1';
                element.style.pointerEvents = 'auto';
            } else {
                element.style.opacity = '0.3';
                element.style.pointerEvents = 'none';
            }
        });
    }

    showConceptDetails(concept) {
        this.activeConcept = concept;

        document.getElementById('conceptTitle').textContent = concept.name;
        document.getElementById('conceptContent').innerHTML = `
            <p><strong>${concept.description}</strong></p>
            ${concept.details}
            <div class="concept-example">
                <h4>Real-World Example</h4>
                <p>${this.getExample(concept.id)}</p>
            </div>
        `;

        // Set difficulty stars
        const stars = '★'.repeat(concept.difficulty) + '☆'.repeat(5 - concept.difficulty);
        document.getElementById('difficultyStars').innerHTML = stars;

        // Set related concepts
        const related = concept.related.map(r =>
            `<span onclick="universe.showConceptDetails(universe.concepts.find(c => c.id === '${r}'))">${r}</span>`
        ).join(', ');
        document.getElementById('relatedConcepts').innerHTML = related;

        document.getElementById('knowledgePanel').classList.add('active');

        // Highlight connections
        this.highlightConnections(concept);
    }

    getExample(conceptId) {
        const examples = {
            'raft': 'Used in etcd for maintaining cluster state in Kubernetes',
            'kafka': 'Used by LinkedIn for activity tracking and operational metrics',
            'consul': 'Used by HashiCorp for service discovery in microservices',
            'cassandra': 'Used by Netflix for storing user viewing data',
            'grpc': 'Used by Google for internal service communication',
            'kubernetes': 'Used by Google, Amazon, and Microsoft for container orchestration'
        };
        return examples[conceptId] || 'Widely used in production distributed systems';
    }

    highlightConnections(concept) {
        document.querySelectorAll('.connection-line').forEach(el => {
            el.classList.remove('active');
        });

        this.connections.forEach(conn => {
            if (conn.from.id === concept.id || conn.to.id === concept.id) {
                const fromEl = conn.from.element;
                const toEl = conn.to.element;

                if (fromEl && toEl) {
                    const line = document.createElement('div');
                    line.className = 'connection-line active';
                    line.style.left = `${conn.from.x}px`;
                    line.style.top = `${conn.from.y}px`;

                    const dx = conn.to.x - conn.from.x;
                    const dy = conn.to.y - conn.from.y;
                    const length = Math.sqrt(dx * dx + dy * dy);
                    const angle = Math.atan2(dy, dx) * 180 / Math.PI;

                    line.style.width = `${length}px`;
                    line.style.transform = `rotate(${angle}deg)`;
                    line.style.transformOrigin = '0 0';

                    document.querySelector('.universe-container').appendChild(line);
                    setTimeout(() => line.remove(), 2000);
                }
            }
        });
    }

    hideConceptDetails() {
        document.getElementById('knowledgePanel').classList.remove('active');
        this.activeConcept = null;
    }

    startGuidedTour() {
        let index = 0;
        const tour = () => {
            if (index < this.concepts.length) {
                this.showConceptDetails(this.concepts[index]);
                setTimeout(tour, 5000);
                index++;
            }
        };
        tour();
    }

    startQuiz() {
        const randomConcept = this.concepts[Math.floor(Math.random() * this.concepts.length)];
        const questions = [
            {
                question: `What is the main purpose of ${randomConcept.name}?`,
                options: [
                    'Data storage',
                    'Service coordination',
                    'Message passing',
                    'Consensus algorithm'
                ],
                answer: 3
            }
        ];

        // Show quiz modal
        this.showQuizModal(questions[0]);
    }

    showQuizModal(question) {
        const modal = document.createElement('div');
        modal.className = 'quiz-modal';
        modal.innerHTML = `
            <div class="quiz-content">
                <h3>Quick Quiz</h3>
                <p>${question.question}</p>
                <div class="quiz-options">
                    ${question.options.map((opt, i) =>
            `<button class="quiz-option" data-index="${i}">${opt}</button>`
        ).join('')}
                </div>
                <button class="close-quiz">Close</button>
            </div>
        `;

        document.body.appendChild(modal);

        modal.querySelectorAll('.quiz-option').forEach(btn => {
            btn.addEventListener('click', () => {
                const isCorrect = parseInt(btn.dataset.index) === question.answer;
                alert(isCorrect ? 'Correct! 🎉' : 'Try again! 💡');
                modal.remove();
            });
        });

        modal.querySelector('.close-quiz').addEventListener('click', () => modal.remove());
    }

    updatePositions() {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;

        this.concepts.forEach(concept => {
            // Orbit around center
            concept.rotation += 0.002 * concept.speed;
            const radius = 300 + concept.z;

            concept.x = centerX + Math.cos(concept.rotation) * radius;
            concept.y = centerY + Math.sin(concept.rotation) * radius;

            // Float animation
            concept.y += Math.sin(this.time * 0.001 + concept.id.length) * 3;

            // Update HTML element position
            if (concept.element) {
                concept.element.style.left = `${concept.x - concept.element.offsetWidth / 2}px`;
                concept.element.style.top = `${concept.y - concept.element.offsetHeight / 2}px`;

                // 3D perspective effect
                const scale = 0.8 + (concept.z + 100) / 400;
                concept.element.style.transform = `scale(${scale})`;
                concept.element.style.zIndex = Math.round(concept.z + 100);
            }
        });

        // Update particles
        this.particles.forEach(p => {
            p.x += p.speed;
            p.y += Math.sin(this.time * 0.001 + p.x * 0.01) * 0.5;

            if (p.x > this.canvas.width) p.x = 0;
            if (p.y > this.canvas.height) p.y = 0;
            if (p.y < 0) p.y = this.canvas.height;
        });
    }

    draw() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Draw gradient background
        const gradient = this.ctx.createRadialGradient(
            this.canvas.width / 2, this.canvas.height / 2, 0,
            this.canvas.width / 2, this.canvas.height / 2, this.canvas.width
        );
        gradient.addColorStop(0, 'rgba(10, 14, 23, 0.1)');
        gradient.addColorStop(1, 'rgba(5, 7, 12, 0.8)');
        this.ctx.fillStyle = gradient;
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

        // Draw particles (stars)
        this.ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
        this.particles.forEach(p => {
            this.ctx.globalAlpha = p.opacity;
            this.ctx.beginPath();
            this.ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
            this.ctx.fill();
        });
        this.ctx.globalAlpha = 1;

        // Draw connections
        if (document.getElementById('showConnections').checked) {
            this.connections.forEach(conn => {
                this.ctx.strokeStyle = 'rgba(100, 200, 255, 0.2)';
                this.ctx.lineWidth = 1;
                this.ctx.beginPath();
                this.ctx.moveTo(conn.from.x, conn.from.y);
                this.ctx.lineTo(conn.to.x, conn.to.y);
                this.ctx.stroke();
            });
        }

        // Draw mouse effect
        const dist = 100;
        this.ctx.strokeStyle = 'rgba(79, 209, 199, 0.1)';
        this.ctx.lineWidth = 1;
        this.ctx.beginPath();
        this.ctx.arc(this.mouse.x, this.mouse.y, dist, 0, Math.PI * 2);
        this.ctx.stroke();
    }

    animate() {
        this.time += 16;
        this.updatePositions();
        this.draw();
        requestAnimationFrame(() => this.animate());
    }
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.universe = new DistributedUniverse();
});
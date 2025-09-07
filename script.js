class ProjectManager {
    constructor() {
        this.projects = [];
        this.certificates = [];
        this.experiences = [];
        this.init();
    }

    async init() {
        await this.loadProjects();
        await this.loadCertificates();
        await this.loadExperiences();
        this.renderProjects();
        this.renderCertificates();
        this.renderExperiences();
    }

    async loadProjects() {
        try {
            const response = await fetch('projects.json');
            this.projects = await response.json();
        } catch (error) {
            console.error('Error loading projects:', error);
            this.projects = [];
        }
    }

    async loadCertificates() {
        try {
            const response = await fetch('certificates.json');
            this.certificates = await response.json();
        } catch (error) {
            console.error('Error loading certificates:', error);
            this.certificates = [];
        }
    }

    async loadExperiences() {
        try {
            const response = await fetch('experiences.json');
            this.experiences = await response.json();
        } catch (error) {
            console.error('Error loading experiences:', error);
            this.experiences = [];
        }
    }

    renderProjects() {
        const grid = document.getElementById('projects-grid');
        
        if (this.projects.length === 0) {
            grid.innerHTML = `
                <div class="col-span-full text-center py-12">
                    <div class="text-zinc-400 text-6xl mb-4">📁</div>
                    <h3 class="text-xl font-medium text-zinc-500 mb-2">No projects available</h3>
                </div>
            `;
            return;
        }

        grid.innerHTML = this.projects.map((project) => `
            <div class="group bg-white/90 backdrop-blur-sm rounded-2xl shadow-lg border border-white/20 p-8 hover:shadow-2xl hover:scale-[1.02] transition-all duration-300 hover:bg-white/95">
                <!-- Header with gradient accent -->
                <div class="flex items-start justify-between mb-6">
                    <div class="flex-1">
                        <h3 class="text-2xl font-bold text-slate-800 mb-2 group-hover:text-blue-600 transition-colors duration-200">${project.name}</h3>
                        <div class="w-16 h-1 bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full"></div>
                    </div>
                </div>
                
                <!-- Description -->
                <p class="text-slate-600 mb-8 leading-relaxed text-base">${project.description}</p>
                
                <!-- Technologies -->
                <div class="mb-8">
                    <h4 class="text-sm font-semibold text-slate-700 mb-3 uppercase tracking-wide">Technologies</h4>
                    <div class="flex flex-wrap gap-2">
                        ${project.technologies.map(tech => 
                            `<span class="bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 text-blue-700 text-sm px-3 py-1.5 rounded-lg font-medium hover:from-blue-100 hover:to-indigo-100 transition-all duration-200">${tech}</span>`
                        ).join('')}
                    </div>
                </div>

                ${project.achievements && project.achievements.length > 0 ? `
                    <!-- Key Achievements -->
                    <div class="mb-8">
                        <h4 class="text-sm font-semibold text-slate-700 mb-4 uppercase tracking-wide">Key Achievements</h4>
                        <div class="space-y-3">
                            ${project.achievements.map(achievement => 
                                `<div class="flex items-start group-hover:translate-x-1 transition-transform duration-200">
                                    <div class="w-2 h-2 bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                                    <p class="text-slate-600 text-sm leading-relaxed">${achievement}</p>
                                </div>`
                            ).join('')}
                        </div>
                    </div>
                ` : ''}
                
                <!-- Action Links -->
                <div class="flex flex-wrap gap-4 pt-4 border-t border-slate-100">
                    ${project.github ? `
                        <a href="${project.github}" target="_blank" 
                           class="inline-flex items-center px-4 py-2.5 bg-gradient-to-r from-slate-800 to-slate-700 text-white rounded-lg hover:from-slate-700 hover:to-slate-600 transition-all duration-200 shadow-md hover:shadow-lg font-medium text-sm">
                            <svg class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 24 24">
                                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                            </svg>
                            View Code
                        </a>
                    ` : ''}
                    ${project.demo ? `
                        <a href="${project.demo}" target="_blank" 
                           class="inline-flex items-center px-4 py-2.5 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg hover:from-blue-600 hover:to-indigo-600 transition-all duration-200 shadow-md hover:shadow-lg font-medium text-sm">
                            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                            </svg>
                            Live Demo
                        </a>
                    ` : ''}
                </div>
            </div>
        `).join('');
    }

    renderCertificates() {
        const grid = document.getElementById('certificates-grid');
        
        if (this.certificates.length === 0) {
            grid.innerHTML = `
                <div class="col-span-full text-center py-12">
                    <div class="text-gray-400 text-6xl mb-4">🏆</div>
                    <h3 class="text-xl font-medium text-gray-300 mb-2">No certificates available</h3>
                </div>
            `;
            return;
        }

        grid.innerHTML = this.certificates.map((certificate) => `
            <div class="group bg-white/90 backdrop-blur-sm rounded-2xl shadow-lg border border-white/20 p-6 hover:shadow-2xl hover:scale-[1.02] transition-all duration-300 hover:bg-white/95 text-center">
                <!-- Certificate Badge/Image -->
                <div class="mb-6">
                    <img src="${certificate.badgeImage}" alt="${certificate.name}" class="w-24 h-24 mx-auto rounded-lg shadow-md">
                </div>
                
                <!-- Certificate Name -->
                <h3 class="text-xl font-bold text-slate-800 mb-2 group-hover:text-amber-600 transition-colors duration-200">${certificate.name}</h3>
                
                <!-- Issuer -->
                <p class="text-slate-600 mb-2 font-medium">${certificate.issuer}</p>
                
                <!-- Date -->
                <p class="text-slate-500 mb-4 text-sm">${certificate.date}</p>
                
                <!-- Description -->
                <p class="text-slate-600 mb-6 text-sm leading-relaxed">${certificate.description}</p>
                
                <!-- View Certificate Link -->
                <div class="pt-4 border-t border-slate-100">
                    <a href="${certificate.credlyUrl}" target="_blank" 
                       class="inline-flex items-center px-4 py-2.5 bg-gradient-to-r from-amber-500 to-yellow-500 text-white rounded-lg hover:from-amber-600 hover:to-yellow-600 transition-all duration-200 shadow-md hover:shadow-lg font-medium text-sm">
                        <svg class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
                        </svg>
                        View Certificate
                    </a>
                </div>
            </div>
        `).join('');
    }

    renderExperiences() {
        const grid = document.getElementById('experience-grid');
        
        if (this.experiences.length === 0) {
            grid.innerHTML = `
                <div class="text-center py-12">
                    <div class="text-gray-400 text-6xl mb-4">💼</div>
                    <h3 class="text-xl font-medium text-gray-300 mb-2">No experience data available</h3>
                </div>
            `;
            return;
        }

        grid.innerHTML = this.experiences.map((experience) => `
            <div class="group bg-white/90 backdrop-blur-sm rounded-2xl shadow-lg border border-white/20 p-8 hover:shadow-2xl transition-all duration-300 hover:bg-white/95">
                <!-- Header -->
                <div class="flex flex-col md:flex-row md:justify-between md:items-start mb-6">
                    <div class="flex-1 mb-4 md:mb-0">
                        <h3 class="text-2xl font-bold text-slate-800 mb-2 group-hover:text-green-600 transition-colors duration-200">${experience.position}</h3>
                        <h4 class="text-xl font-semibold text-slate-600 mb-2">${experience.company}</h4>
                        <p class="text-slate-500 text-sm">${experience.location}</p>
                    </div>
                    <div class="flex flex-col items-start md:items-end">
                        <span class="text-slate-600 font-medium mb-2">${experience.period}</span>
                        <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ${
                            experience.type === 'Current' ? 'bg-green-100 text-green-800' :
                            experience.type === 'Contract' ? 'bg-blue-100 text-blue-800' :
                            'bg-gray-100 text-gray-800'
                        }">${experience.type}</span>
                    </div>
                </div>

                <!-- Description -->
                <p class="text-slate-600 mb-6 leading-relaxed">${experience.description}</p>

                <!-- Key Highlights -->
                <div class="mb-6">
                    <h5 class="text-sm font-semibold text-slate-700 mb-3 uppercase tracking-wide">Key Highlights</h5>
                    <div class="space-y-2">
                        ${experience.highlights.map(highlight => `
                            <div class="flex items-start">
                                <div class="w-2 h-2 bg-gradient-to-r from-green-500 to-teal-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                                <p class="text-slate-600 text-sm leading-relaxed">${highlight}</p>
                            </div>
                        `).join('')}
                    </div>
                </div>

                <!-- Technologies -->
                <div>
                    <h5 class="text-sm font-semibold text-slate-700 mb-3 uppercase tracking-wide">Technologies Used</h5>
                    <div class="flex flex-wrap gap-2">
                        ${experience.technologies.map(tech => 
                            `<span class="bg-gradient-to-r from-green-50 to-teal-50 border border-green-200 text-green-700 text-sm px-3 py-1.5 rounded-lg font-medium hover:from-green-100 hover:to-teal-100 transition-all duration-200">${tech}</span>`
                        ).join('')}
                    </div>
                </div>
            </div>
        `).join('');
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new ProjectManager();
});
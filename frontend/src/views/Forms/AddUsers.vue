<template>
  <AdminLayout>
    <PageBreadcrumb :pageTitle="currentPageTitle" />
    
    <form @submit.prevent="handleSubmit" class="space-y-6">
      
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
        
        <div class="space-y-6">
          <ComponentCard title="Dados de Login e Identificação">
            <DefaultInputs v-model="form" />
          </ComponentCard>
          
          <ComponentCard title="Função e Permissões">
            <SelectInput v-model="form.role" />
          </ComponentCard>

          <ComponentCard title="Biografia / Observações">
            <TextArea v-model="form.description" />
          </ComponentCard>
        </div>
        
        <div class="space-y-6">
          <ComponentCard title="Contactos Oficiais">
            <InputGroup 
              v-model:email="form.contactEmail"
              v-model:phone="form.phone"
              v-model:country="form.country"
            />
          </ComponentCard>
          
          <ComponentCard title="Fotografia de Perfil">
            <FileInput v-model="form.avatar" />
          </ComponentCard>
          
          <ComponentCard title="Definições de Conta">
            <CheckboxInput v-model="form.isActive" />
          </ComponentCard>
        </div>

      </div>

      <div class="flex justify-end gap-4 rounded-xl border border-gray-200 bg-white p-4 dark:border-gray-800 dark:bg-white/[0.03]">
        <button 
          type="button" 
          @click="resetForm"
          class="px-5 py-2.5 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
        >
          Cancelar
        </button>
        <button 
          type="submit"
          class="px-5 py-2.5 text-sm font-medium text-white bg-brand-500 hover:bg-brand-600 rounded-lg shadow-theme-xs transition-colors dark:bg-brand-600 dark:hover:bg-brand-700"
        >
          Criar Utilizador
        </button>
      </div>

    </form>
  </AdminLayout>
</template>

<script setup>
import { ref, reactive } from 'vue'
import PageBreadcrumb from '@/components/common/PageBreadcrumb.vue'
import AdminLayout from '@/components/layout/AdminLayout.vue'
import ComponentCard from '@/components/common/ComponentCard.vue'

// Importações dos teus componentes de inputs limpos
import DefaultInputs from '@/components/forms/FormElements/DefaultInputs.vue'
import SelectInput from '@/components/forms/FormElements/SelectInput.vue'
import TextArea from '@/components/forms/FormElements/TextArea.vue'
import InputGroup from '@/components/forms/FormElements/InputGroup.vue'
import FileInput from '@/components/forms/FormElements/FileInput.vue'
import CheckboxInput from '@/components/forms/FormElements/CheckboxInput.vue'

const currentPageTitle = ref('Criar Novo Utilizador')

// Estado único reativo que junta tudo o que vais enviar para a tua API/Backend
const form = reactive({
  name: '',
  password: '',
  role: '',
  description: '',
  contactEmail: '',
  phone: '',
  country: 'US',
  avatar: null,
  isActive: true
})

const handleSubmit = () => {
  console.log('Dados prontos para enviar para o Servidor:', { ...form })
  alert(`Utilizador ${form.name} criado com sucesso! (Ver consola)`)
}

const resetForm = () => {
  form.name = ''
  form.password = ''
  form.role = ''
  form.description = ''
  form.contactEmail = ''
  form.phone = ''
  form.country = 'US'
  form.avatar = null
  form.isActive = true
}
</script>
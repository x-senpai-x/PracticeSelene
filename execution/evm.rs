let evm = Revm::builder().with_db(db).with_env(env).build();//creates instance of evm with env variables and db provided 
let mut ctx: revm::ContextWithHandlerCfg<(), ProofDB<R>> = evm.into_context_with_handler_cfg();
let mut evm = Revm::builder().with_context_with_handler_cfg(ctx).build(); //rebuilds evm with updated context
let res: std::result::Result<ResultAndState, revm::primitives::EVMError<Report>> = evm.transact();//executes transaction ; res is success or failure 
convert to go lang
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug, Eq, PartialEq)]
#[non_exhaustive]
//Implemnt with_db
//First complete innerEvmContext
impl<'a, EXT, DB: Database> EvmBuilder<'a, SetGenericStage, EXT, DB> {

pub fn with_db<ODB: Database>(self, db: ODB) -> EvmBuilder<'a, SetGenericStage, EXT, ODB> {
    EvmBuilder {
        context: Context::new(self.context.evm.with_db(db), self.context.external),
        handler: EvmBuilder::<'a, SetGenericStage, EXT, ODB>::handler(self.handler.cfg()),
        phantom: PhantomData,
    }
}
impl<DB: Database> EvmContext<DB> {

pub fn with_db<ODB: Database>(self, db: ODB) -> EvmContext<ODB> {
    EvmContext {
        inner: self.inner.with_db(db),
        precompiles: ContextPrecompiles::default(),
    }
}
impl<DB: Database> InnerEvmContext<DB> {

pub fn with_db<ODB: Database>(self, db: ODB) -> InnerEvmContext<ODB> {
    InnerEvmContext {
        env: self.env,
        journaled_state: self.journaled_state,
        db,
        error: Ok(()),
        valid_authorizations: Default::default(),
        #[cfg(feature = "optimism")]
        l1_block_info: self.l1_block_info,
    }
}
impl<DB: Database> Default for ContextPrecompiles<DB> {
    fn default() -> Self {
        Self {
            inner: Default::default(),
        }
    }
}
impl<DB: Database> Default for PrecompilesCow<DB> {
    fn default() -> Self {
        Self::Owned(Default::default())
    }
}
impl<DB: Database> EvmContext<DB> {
    /// Create new context with database.
    pub fn new(db: DB) -> Self {
        Self {
            inner: InnerEvmContext::new(db),
            precompiles: ContextPrecompiles::default(),
        }
    }
    impl<DB: Database> Default for ContextPrecompiles<DB> {
        fn default() -> Self {
            Self {
                inner: Default::default(),
            }
        }
    }
    impl<DB: Database> Default for PrecompilesCow<DB> {
        fn default() -> Self {
            Self::Owned(Default::default())
        }
    }
    #[stable(feature = "rust1", since = "1.0.0")]
impl<K, V, S> Default for HashMap<K, V, S>
where
    S: Default,
{
    /// Creates an empty `HashMap<K, V, S>`, with the `Default` value for the hasher.
    #[inline]
    fn default() -> HashMap<K, V, S> {
        HashMap::with_hasher(Default::default())
    }
}
pub struct CfgEnv {
    /// Chain ID of the EVM, it will be compared to the transaction's Chain ID.
    /// Chain ID is introduced EIP-155
    pub chain_id: u64,
    /// KZG Settings for point evaluation precompile. By default, this is loaded from the ethereum mainnet trusted setup.
    #[cfg(any(feature = "c-kzg", feature = "kzg-rs"))]
    #[cfg_attr(feature = "serde", serde(skip))]
    pub kzg_settings: crate::kzg::EnvKzgSettings,
    /// Bytecode that is created with CREATE/CREATE2 is by default analysed and jumptable is created.
    /// This is very beneficial for testing and speeds up execution of that bytecode if called multiple times.
    ///
    /// Default: Analyse
    pub perf_analyse_created_bytecodes: AnalysisKind,
    /// If some it will effects EIP-170: Contract code size limit. Useful to increase this because of tests.
    /// By default it is 0x6000 (~25kb).
    pub limit_contract_code_size: Option<usize>,
    /// A hard memory limit in bytes beyond which [crate::result::OutOfGasError::Memory] cannot be resized.
    ///
    /// In cases where the gas limit may be extraordinarily high, it is recommended to set this to
    /// a sane value to prevent memory allocation panics. Defaults to `2^32 - 1` bytes per
    /// EIP-1985.
    #[cfg(feature = "memory_limit")]
    pub memory_limit: u64,
    /// Skip balance checks if true. Adds transaction cost to balance to ensure execution doesn't fail.
    #[cfg(feature = "optional_balance_check")]
    pub disable_balance_check: bool,
    /// There are use cases where it's allowed to provide a gas limit that's higher than a block's gas limit. To that
    /// end, you can disable the block gas limit validation.
    /// By default, it is set to `false`.
    #[cfg(feature = "optional_block_gas_limit")]
    pub disable_block_gas_limit: bool,
    /// EIP-3607 rejects transactions from senders with deployed code. In development, it can be desirable to simulate
    /// calls from contracts, which this setting allows.
    /// By default, it is set to `false`.
    #[cfg(feature = "optional_eip3607")]
    pub disable_eip3607: bool,
    /// Disables all gas refunds. This is useful when using chains that have gas refunds disabled e.g. Avalanche.
    /// Reasoning behind removing gas refunds can be found in EIP-3298.
    /// By default, it is set to `false`.
    #[cfg(feature = "optional_gas_refund")]
    pub disable_gas_refund: bool,
    /// Disables base fee checks for EIP-1559 transactions.
    /// This is useful for testing method calls with zero gas price.
    /// By default, it is set to `false`.
    #[cfg(feature = "optional_no_base_fee")]
    pub disable_base_fee: bool,
    /// Disables the payout of the reward to the beneficiary.
    /// By default, it is set to `false`.
    #[cfg(feature = "optional_beneficiary_reward")]
    pub disable_beneficiary_reward: bool,
}
        pub enum EnvKzgSettings {
            /// Default mainnet trusted setup
            #[default]
            Default,
            /// Custom trusted setup.
            Custom(std::sync::Arc<c_kzg::KzgSettings>),
        }
pub enum AnalysisKind {
    Raw,
    Analyse,
}
//Defining every field of struct even if not in use?
//Implementing transact 
//Implementing builder
//Implementing handler
//Doubt here
pub struct ValidationHandler<'a, EXT, DB: Database> {
    /// Validate and calculate initial transaction gas.
    pub initial_tx_gas: ValidateInitialTxGasHandle<'a, DB>,
    /// Validate transactions against state data.
    pub tx_against_state: ValidateTxEnvAgainstState<'a, EXT, DB>,
    /// Validate Env.
    pub env: ValidateEnvHandle<'a, DB>,
}
pub type ValidateEnvHandle<'a, DB> =
    Arc<dyn Fn(&Env) -> Result<(), EVMError<<DB as Database>::Error>> + 'a>;

/// Handle that validates transaction environment against the state.
/// Second parametar is initial gas.
pub type ValidateTxEnvAgainstState<'a, EXT, DB> =
    Arc<dyn Fn(&mut Context<EXT, DB>) -> Result<(), EVMError<<DB as Database>::Error>> + 'a>;

/// Initial gas calculation handle
pub type ValidateInitialTxGasHandle<'a, DB> =
    Arc<dyn Fn(&Env) -> Result<u64, EVMError<<DB as Database>::Error>> + 'a>;
impl<'a> Evm<'a, (), EmptyDB> {
        /// Returns evm builder with empty database and empty external context.
    pub fn builder() -> EvmBuilder<'a, SetGenericStage, (), EmptyDB> {
       EvmBuilder::default()        }
}

pub struct EvmBuilder<'a, BuilderStage, EXT, DB: Database> {
    context: Context<EXT, DB>,
    /// Handler that will be used by EVM. It contains handle registers
    handler: Handler<'a, Context<EXT, DB>, EXT, DB>,
    /// Phantom data to mark the stage of the builder.
    phantom: PhantomData<BuilderStage>,
}
impl<'a> Default for EvmBuilder<'a, SetGenericStage, (), EmptyDB> {
    fn default() -> Self {
        cfg_if::cfg_if! {
            if #[cfg(all(feature = "optimism-default-handler",
                not(feature = "negate-optimism-default-handler")))] {
                    let mut handler_cfg = HandlerCfg::new(SpecId::LATEST);
                    // set is_optimism to true by default.
                    handler_cfg.is_optimism = true;

            } else {
                let handler_cfg = HandlerCfg::new(SpecId::LATEST);
            }
        }

        Self {
            context: Context::default(),
            handler: EvmBuilder::<'a, SetGenericStage, (), EmptyDB>::handler(handler_cfg),
            phantom: PhantomData,
        }
    }
}
pub struct Evm<'a, EXT, DB: Database> {
    /// Context of execution, containing both EVM and external context.
    pub context: Context<EXT, DB>,
    /// Handler is a component of the of EVM that contains all the logic. Handler contains specification id
    /// and it different depending on the specified fork.
    pub handler: Handler<'a, Context<EXT, DB>, EXT, DB>,
}
pub struct Context<EXT, DB: Database> {
    /// Evm Context (internal context).
    pub evm: EvmContext<DB>,
    /// External contexts.
    pub external: EXT,
}
pub struct Handler<'a, H: Host + 'a, EXT, DB: Database> {
    /// Handler configuration.
    pub cfg: HandlerCfg,
    /// Instruction table type.
    pub instruction_table: InstructionTables<'a, H>,
    /// Registers that will be called on initialization.
    pub registers: Vec<HandleRegisters<'a, EXT, DB>>,
    /// Validity handles.
    pub validation: ValidationHandler<'a, EXT, DB>,
    /// Pre execution handle.
    pub pre_execution: PreExecutionHandler<'a, EXT, DB>,
    /// Post Execution handle.
    pub post_execution: PostExecutionHandler<'a, EXT, DB>,
    /// Execution loop that handles frames.
    pub execution: ExecutionHandler<'a, EXT, DB>,
}
pub struct EvmContext<DB: Database> {
    /// Inner EVM context.
    pub inner: InnerEvmContext<DB>,
    /// Precompiles that are available for evm.
    pub precompiles: ContextPrecompiles<DB>,
}
pub struct ContextPrecompiles<DB: Database> {
    inner: PrecompilesCow<DB>,
}
enum PrecompilesCow<DB: Database> {
    /// Default precompiles, returned by `Precompiles::new`. Used to fast-path the default case.
    StaticRef(&'static Precompiles),
    Owned(HashMap<Address, ContextPrecompile<DB>>),
}
pub enum ContextPrecompile<DB: Database> {
    /// Ordinary precompiles
    Ordinary(Precompile),
    /// Stateful precompile that is Arc over [`ContextStatefulPrecompile`] trait.
    /// It takes a reference to input, gas limit and Context.
    ContextStateful(ContextStatefulPrecompileArc<DB>),
    /// Mutable stateful precompile that is Box over [`ContextStatefulPrecompileMut`] trait.
    /// It takes a reference to input, gas limit and context.
    ContextStatefulMut(ContextStatefulPrecompileBox<DB>),
}
pub type ContextStatefulPrecompileArc<DB> = Arc<dyn ContextStatefulPrecompile<DB>>;

/// Box over context mutable stateful precompile
pub type ContextStatefulPrecompileBox<DB> = Box<dyn ContextStatefulPrecompileMut<DB>>;
pub struct Precompiles {
    /// Precompiles.
    inner: HashMap<Address, Precompile>,
    /// Addresses of precompile.
    addresses: HashSet<Address>,
}
pub enum Precompile {
    /// Standard simple precompile that takes input and gas limit.
    Standard(StandardPrecompileFn),
    /// Similar to Standard but takes reference to environment.
    Env(EnvPrecompileFn),
    /// Stateful precompile that is Arc over [`StatefulPrecompile`] trait.
    /// It takes a reference to input, gas limit and environment.
    Stateful(StatefulPrecompileArc),
    /// Mutable stateful precompile that is Box over [`StatefulPrecompileMut`] trait.
    /// It takes a reference to input, gas limit and environment.
    StatefulMut(StatefulPrecompileBox),
}
pub type StandardPrecompileFn = fn(&Bytes, u64) -> PrecompileResult;
pub type EnvPrecompileFn = fn(&Bytes, u64, env: &Env) -> PrecompileResult;

pub trait StatefulPrecompile: Sync + Send {
    fn call(&self, bytes: &Bytes, gas_limit: u64, env: &Env) -> PrecompileResult;
}

/// Mutable stateful precompile trait. It is used to create
/// a boxed precompile in Precompile::StatefulMut.
pub trait StatefulPrecompileMut: DynClone + Send + Sync {
    fn call_mut(&mut self, bytes: &Bytes, gas_limit: u64, env: &Env) -> PrecompileResult;
}

dyn_clone::clone_trait_object!(StatefulPrecompileMut);

/// Arc over stateful precompile.
pub type StatefulPrecompileArc = Arc<dyn StatefulPrecompile>;

/// Box over mutable stateful precompile
pub type StatefulPrecompileBox = Box<dyn StatefulPrecompileMut>;
pub type PrecompileResult = Result<PrecompileOutput, PrecompileErrors>;
pub struct PrecompileOutput {
    /// Gas used by the precompile.
    pub gas_used: u64,
    /// Output bytes.
    pub bytes: Bytes,
}
pub enum PrecompileErrors {
    Error(PrecompileError),
    Fatal { msg: String },
}
pub enum PrecompileError {
    /// out of gas is the main error. Others are here just for completeness
    OutOfGas,
    // Blake2 errors
    Blake2WrongLength,
    Blake2WrongFinalIndicatorFlag,
    // Modexp errors
    ModexpExpOverflow,
    ModexpBaseOverflow,
    ModexpModOverflow,
    // Bn128 errors
    Bn128FieldPointNotAMember,
    Bn128AffineGFailedToCreate,
    Bn128PairLength,
    // Blob errors
    /// The input length is not exactly 192 bytes.
    BlobInvalidInputLength,
    /// The commitment does not match the versioned hash.
    BlobMismatchedVersion,
    /// The proof verification failed.
    BlobVerifyKzgProofFailed,
    /// Catch-all variant for other errors.
    Other(String),
}

pub struct InnerEvmContext<DB: Database> {
    /// EVM Environment contains all the information about config, block and transaction that
    /// evm needs.
    pub env: Box<Env>,
    /// EVM State with journaling support.
    pub journaled_state: JournaledState,
    /// Database to load data from.
    pub db: DB,
    /// Error that happened during execution.
    pub error: Result<(), EVMError<DB::Error>>,
    /// EIP-7702 Authorization list of accounts that needs to be cleared.
    pub valid_authorizations: Vec<Address>,
    /// Used as temporary value holder to store L1 block info.
    #[cfg(feature = "optimism")]
    pub l1_block_info: Option<crate::optimism::L1BlockInfo>,
}
pub struct JournaledState {
    /// Current state.
    pub state: EvmState,
    /// [EIP-1153](https://eips.ethereum.org/EIPS/eip-1153) transient storage that is discarded after every transactions
    pub transient_storage: TransientStorage,
    /// logs
    pub logs: Vec<Log>,
    /// how deep are we in call stack.
    pub depth: usize,
    /// journal with changes that happened between calls.
    pub journal: Vec<Vec<JournalEntry>>,
    /// Ethereum before EIP-161 differently defined empty and not-existing account
    /// Spec is needed for two things SpuriousDragon's `EIP-161 State clear`,
    /// and for Cancun's `EIP-6780: SELFDESTRUCT in same transaction`
    pub spec: SpecId,
    /// Warm loaded addresses are used to check if loaded address
    /// should be considered cold or warm loaded when the account
    /// is first accessed.
    ///
    /// Note that this not include newly loaded accounts, account and storage
    /// is considered warm if it is found in the `State`.
    pub warm_preloaded_addresses: HashSet<Address>,
}
pub enum JournalEntry {
    /// Used to mark account that is warm inside EVM in regards to EIP-2929 AccessList.
    /// Action: We will add Account to state.
    /// Revert: we will remove account from state.
    AccountWarmed { address: Address },
    /// Mark account to be destroyed and journal balance to be reverted
    /// Action: Mark account and transfer the balance
    /// Revert: Unmark the account and transfer balance back
    AccountDestroyed {
        address: Address,
        target: Address,
        was_destroyed: bool, // if account had already been destroyed before this journal entry
        had_balance: U256,
    },
    /// Loading account does not mean that account will need to be added to MerkleTree (touched).
    /// Only when account is called (to execute contract or transfer balance) only then account is made touched.
    /// Action: Mark account touched
    /// Revert: Unmark account touched
    AccountTouched { address: Address },
    /// Transfer balance between two accounts
    /// Action: Transfer balance
    /// Revert: Transfer balance back
    BalanceTransfer {
        from: Address,
        to: Address,
        balance: U256,
    },
    /// Increment nonce
    /// Action: Increment nonce by one
    /// Revert: Decrement nonce by one
    NonceChange {
        address: Address, //geth has nonce value,
    },
    /// Create account:
    /// Actions: Mark account as created
    /// Revert: Unmart account as created and reset nonce to zero.
    AccountCreated { address: Address },
    /// Entry used to track storage changes
    /// Action: Storage change
    /// Revert: Revert to previous value
    StorageChanged {
        address: Address,
        key: U256,
        had_value: U256,
    },
    /// Entry used to track storage warming introduced by EIP-2929.
    /// Action: Storage warmed
    /// Revert: Revert to cold state
    StorageWarmed { address: Address, key: U256 },
    /// It is used to track an EIP-1153 transient storage change.
    /// Action: Transient storage changed.
    /// Revert: Revert to previous value.
    TransientStorageChange {
        address: Address,
        key: U256,
        had_value: U256,
    },
    /// Code changed
    /// Action: Account code changed
    /// Revert: Revert to previous bytecode.
    CodeChange { address: Address },
}

pub struct HandlerCfg {
    /// Specification identification.
    pub spec_id: SpecId,
    /// Optimism related field, it will append the Optimism handle register to the EVM.
    #[cfg(feature = "optimism")]
    pub is_optimism: bool,
}
pub enum InstructionTables<'a, H: ?Sized> {
    Plain(InstructionTable<H>),
    Boxed(BoxedInstructionTable<'a, H>),
}
pub type Instruction<H> = fn(&mut Interpreter, &mut H);

/// Instruction table is list of instruction function pointers mapped to 256 EVM opcodes.
pub type InstructionTable<H> = [Instruction<H>; 256];
pub struct ValidationHandler<'a, EXT, DB: Database> {
    /// Validate and calculate initial transaction gas.
    pub initial_tx_gas: ValidateInitialTxGasHandle<'a, DB>,
    /// Validate transactions against state data.
    pub tx_against_state: ValidateTxEnvAgainstState<'a, EXT, DB>,
    /// Validate Env.
    pub env: ValidateEnvHandle<'a, DB>,
}
pub type ValidateEnvHandle<'a, DB> =
    Arc<dyn Fn(&Env) -> Result<(), EVMError<<DB as Database>::Error>> + 'a>;

/// Handle that validates transaction environment against the state.
/// Second parametar is initial gas.
pub type ValidateTxEnvAgainstState<'a, EXT, DB> =
    Arc<dyn Fn(&mut Context<EXT, DB>) -> Result<(), EVMError<<DB as Database>::Error>> + 'a>;

/// Initial gas calculation handle
pub type ValidateInitialTxGasHandle<'a, DB> =
    Arc<dyn Fn(&Env) -> Result<u64, EVMError<<DB as Database>::Error>> + 'a>;
pub struct PreExecutionHandler<'a, EXT, DB: Database> {
    /// Load precompiles
    pub load_precompiles: LoadPrecompilesHandle<'a, DB>,
    /// Main load handle
    pub load_accounts: LoadAccountsHandle<'a, EXT, DB>,
    /// Deduct max value from the caller.
    pub deduct_caller: DeductCallerHandle<'a, EXT, DB>,
}
pub type LoadPrecompilesHandle<'a, DB> = Arc<dyn Fn() -> ContextPrecompiles<DB> + 'a>;

/// Load access list accounts and beneficiary.
/// There is no need to load Caller as it is assumed that
/// it will be loaded in DeductCallerHandle.
pub type LoadAccountsHandle<'a, EXT, DB> =
    Arc<dyn Fn(&mut Context<EXT, DB>) -> Result<(), EVMError<<DB as Database>::Error>> + 'a>;

/// Deduct the caller to its limit.
pub type DeductCallerHandle<'a, EXT, DB> =
    Arc<dyn Fn(&mut Context<EXT, DB>) -> EVMResultGeneric<(), <DB as Database>::Error> + 'a>;
pub struct PostExecutionHandler<'a, EXT, DB: Database> {
    /// Reimburse the caller with ethereum it didn't spent.
    pub reimburse_caller: ReimburseCallerHandle<'a, EXT, DB>,
    /// Reward the beneficiary with caller fee.
    pub reward_beneficiary: RewardBeneficiaryHandle<'a, EXT, DB>,
    /// Main return handle, returns the output of the transact.
    pub output: OutputHandle<'a, EXT, DB>,
    /// Called when execution ends.
    /// End handle in comparison to output handle will be called every time after execution.
    /// Output in case of error will not be called.
    pub end: EndHandle<'a, EXT, DB>,
    /// Clear handle will be called always. In comparison to end that
    /// is called only on execution end, clear handle is called even if validation fails.
    pub clear: ClearHandle<'a, EXT, DB>,
}
pub type ReimburseCallerHandle<'a, EXT, DB> =
    Arc<dyn Fn(&mut Context<EXT, DB>, &Gas) -> EVMResultGeneric<(), <DB as Database>::Error> + 'a>;

/// Reward beneficiary with transaction rewards.
pub type RewardBeneficiaryHandle<'a, EXT, DB> = ReimburseCallerHandle<'a, EXT, DB>;

/// Main return handle, takes state from journal and transforms internal result to external.
pub type OutputHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            FrameResult,
        ) -> Result<ResultAndState, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// End handle, takes result and state and returns final result.
/// This will be called after all the other handlers.
///
/// It is useful for catching errors and returning them in a different way.
pub type EndHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            Result<ResultAndState, EVMError<<DB as Database>::Error>>,
        ) -> Result<ResultAndState, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Clear handle, doesn't have output, its purpose is to clear the
/// context. It will be always called even on failed validation.
pub type ClearHandle<'a, EXT, DB> = Arc<dyn Fn(&mut Context<EXT, DB>) + 'a>;
pub struct ExecutionHandler<'a, EXT, DB: Database> {
    /// Handles last frame return, modified gas for refund and
    /// sets tx gas limit.
    pub last_frame_return: LastFrameReturnHandle<'a, EXT, DB>,
    /// Executes a single frame.
    pub execute_frame: ExecuteFrameHandle<'a, EXT, DB>,
    /// Frame call
    pub call: FrameCallHandle<'a, EXT, DB>,
    /// Call return
    pub call_return: FrameCallReturnHandle<'a, EXT, DB>,
    /// Insert call outcome
    pub insert_call_outcome: InsertCallOutcomeHandle<'a, EXT, DB>,
    /// Frame crate
    pub create: FrameCreateHandle<'a, EXT, DB>,
    /// Crate return
    pub create_return: FrameCreateReturnHandle<'a, EXT, DB>,
    /// Insert create outcome.
    pub insert_create_outcome: InsertCreateOutcomeHandle<'a, EXT, DB>,
    /// Frame EOFCreate
    pub eofcreate: FrameEOFCreateHandle<'a, EXT, DB>,
    /// EOFCreate return
    pub eofcreate_return: FrameEOFCreateReturnHandle<'a, EXT, DB>,
    /// Insert EOFCreate outcome.
    pub insert_eofcreate_outcome: InsertEOFCreateOutcomeHandle<'a, EXT, DB>,
}
pub type LastFrameReturnHandle<'a, EXT, DB> = Arc<
    dyn Fn(&mut Context<EXT, DB>, &mut FrameResult) -> Result<(), EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Executes a single frame. Errors can be returned in the EVM context.
pub type ExecuteFrameHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Frame,
            &mut SharedMemory,
            &InstructionTables<'_, Context<EXT, DB>>,
            &mut Context<EXT, DB>,
        ) -> Result<InterpreterAction, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Handle sub call.
pub type FrameCallHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            Box<CallInputs>,
        ) -> Result<FrameOrResult, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Handle call return
pub type FrameCallReturnHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            Box<CallFrame>,
            InterpreterResult,
        ) -> Result<CallOutcome, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Insert call outcome to the parent
pub type InsertCallOutcomeHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            &mut Frame,
            &mut SharedMemory,
            CallOutcome,
        ) -> Result<(), EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Handle sub create.
pub type FrameCreateHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            Box<CreateInputs>,
        ) -> Result<FrameOrResult, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Handle create return
pub type FrameCreateReturnHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            Box<CreateFrame>,
            InterpreterResult,
        ) -> Result<CreateOutcome, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Insert call outcome to the parent
pub type InsertCreateOutcomeHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            &mut Frame,
            CreateOutcome,
        ) -> Result<(), EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Handle EOF sub create.
pub type FrameEOFCreateHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            Box<EOFCreateInputs>,
        ) -> Result<FrameOrResult, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Handle EOF create return
pub type FrameEOFCreateReturnHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            Box<EOFCreateFrame>,
            InterpreterResult,
        ) -> Result<CreateOutcome, EVMError<<DB as Database>::Error>>
        + 'a,
>;

/// Insert EOF crate outcome to the parent
pub type InsertEOFCreateOutcomeHandle<'a, EXT, DB> = Arc<
    dyn Fn(
            &mut Context<EXT, DB>,
            &mut Frame,
            CreateOutcome,
        ) -> Result<(), EVMError<<DB as Database>::Error>>
        + 'a,
>;

impl Default for Context<(), EmptyDB> {
    fn default() -> Self {
        Self::new_empty()
    }
}



fn handler(handler_cfg: HandlerCfg) -> Handler<'a, Context<EXT, DB>, EXT, DB> {
    Handler::new(handler_cfg)
}
pub enum SpecId {
    FRONTIER = 0,         // Frontier               0
    FRONTIER_THAWING = 1, // Frontier Thawing       200000
    HOMESTEAD = 2,        // Homestead              1150000
    DAO_FORK = 3,         // DAO Fork               1920000
    TANGERINE = 4,        // Tangerine Whistle      2463000
    SPURIOUS_DRAGON = 5,  // Spurious Dragon        2675000
    BYZANTIUM = 6,        // Byzantium              4370000
    CONSTANTINOPLE = 7,   // Constantinople         7280000 is overwritten with PETERSBURG
    PETERSBURG = 8,       // Petersburg             7280000
    ISTANBUL = 9,         // Istanbul	            9069000
    MUIR_GLACIER = 10,    // Muir Glacier           9200000
    BERLIN = 11,          // Berlin	                12244000
    LONDON = 12,          // London	                12965000
    ARROW_GLACIER = 13,   // Arrow Glacier          13773000
    GRAY_GLACIER = 14,    // Gray Glacier           15050000
    MERGE = 15,           // Paris/Merge            15537394 (TTD: 58750000000000000000000)
    SHANGHAI = 16,        // Shanghai               17034870 (Timestamp: 1681338455)
    CANCUN = 17,          // Cancun                 19426587 (Timestamp: 1710338135)
    PRAGUE = 18,          // Praque                 TBD
    PRAGUE_EOF = 19,      // Praque+EOF             TBD
    #[default]
    LATEST = u8::MAX,
}
pub struct PhantomData<T: ?Sized>;


